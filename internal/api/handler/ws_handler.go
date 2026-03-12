package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go-chat/internal/model"
	"go-chat/internal/service"
	"go-chat/internal/ws"
)

type WSHandler struct {
	authService *service.AuthService
	chatService *service.ChatService
	manager     *ws.Manager
	redisClient *redis.Client
	upgrader    websocket.Upgrader
}

func NewWSHandler(authService *service.AuthService, chatService *service.ChatService, manager *ws.Manager, redisClient *redis.Client) *WSHandler {
	return &WSHandler{
		authService: authService,
		chatService: chatService,
		manager:     manager,
		redisClient: redisClient,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}
}

func (h *WSHandler) Handle(c *gin.Context) {
	rawToken := strings.TrimSpace(c.Query("token")) // 从网址里拿 Token，如 ?token=xxx
	if rawToken == "" {
		writeError(c, http.StatusUnauthorized, "missing token")
		return
	}

	claims, err := h.authService.ParseToken(rawToken) //// 解析 Token，看看你是谁
	if err != nil {
		writeError(c, http.StatusUnauthorized, "invalid token")
		return
	}

	//实时长连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	sessionID := fmt.Sprintf("%d-%d", claims.UserID, time.Now().UnixNano())

	client := ws.NewClient(claims.UserID, conn, func(_ uint64) {
		_, _ = h.setOnline(claims.UserID, sessionID, true)
	}) // 创建一个“客户端”对象
	h.manager.Register(client)
	becameOnline, _ := h.setOnline(claims.UserID, sessionID, true) // 在 Redis 里标记：我上线啦！
	if becameOnline {
		h.broadcastOnlineStatus(claims.UserID, true, "")
	}
	defer func() {
		becameOffline, lastSeenAt := h.setOnline(claims.UserID, sessionID, false)
		h.manager.Unregister(client)
		if becameOffline {
			h.broadcastOnlineStatus(claims.UserID, false, lastSeenAt)
		}
	}()

	go client.WritePump()

	client.ReadPump(func(incoming ws.IncomingEnvelope) {
		if incoming.Type != "chat" {
			client.SendError("unsupported message type")
			return
		}

		// 把消息存入数据库（单聊或群聊）
		message, recipients, err := h.dispatchChat(context.Background(), claims.UserID, incoming.Payload)
		if err != nil {
			client.SendError(err.Error())
			return
		}

		sender, err := h.authService.GetUserByID(context.Background(), claims.UserID)
		if err != nil {
			client.SendError("failed to load sender profile")
			return
		}

		// 把消息包装成一个标准的格式（包含发送者头像、名字、内容、时间）
		outbound := ws.OutgoingEnvelope{
			Type: "message",
			Payload: ws.OutgoingChatPayload{
				ID:        message.ID,
				Sender:    ws.SenderPayload{ID: sender.ID, Username: sender.Username, Avatar: sender.Avatar},
				TargetID:  incoming.Payload.TargetID,
				ChatType:  incoming.Payload.ChatType,
				MsgType:   message.MsgType,
				Content:   message.Content,
				CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
			},
		}

		// 转成 JSON 字符串
		payload, err := json.Marshal(outbound)
		if err != nil {
			client.SendError("failed to encode message")
			return
		}

		for _, userID := range recipients {
			// Using Broadcast to handle both local and cross-node delivery
			_ = h.manager.Broadcast(userID, payload)
		}
	})
}

func (h *WSHandler) dispatchChat(ctx context.Context, senderID uint64, payload ws.IncomingChatPayload) (*model.Message, []uint64, error) {
	switch payload.ChatType {
	case "private":
		return h.chatService.SendPrivateMessage(ctx, senderID, payload.TargetID, payload.MsgType, payload.Content)
	case "group":
		return h.chatService.SendGroupMessage(ctx, senderID, payload.TargetID, payload.MsgType, payload.Content)
	default:
		return nil, nil, service.ErrInvalidChatType
	}
}

func (h *WSHandler) setOnline(userID uint64, sessionID string, online bool) (bool, string) {
	if h.redisClient == nil {
		return false, ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	onlineKey := service.UserOnlineKey(userID)
	lastSeenKey := service.UserLastSeenKey(userID)
	nowText := time.Now().UTC().Format(time.RFC3339)

	if online {
		_, err := h.redisClient.Get(ctx, onlineKey).Result()
		wasOnline := err == nil
		if err != nil && err != redis.Nil {
			return false, ""
		}
		if err := h.redisClient.Set(ctx, onlineKey, sessionID, service.OnlineTTL).Err(); err != nil {
			return false, ""
		}
		return !wasOnline, ""
	}

	// 仅当 Redis 中的会话与当前连接一致时才删除，避免快速重连被旧连接误删在线键
	const releaseScript = `
local current = redis.call("GET", KEYS[1])
if current == ARGV[1] then
  redis.call("DEL", KEYS[1])
  redis.call("SET", KEYS[2], ARGV[2], "EX", ARGV[3])
  return 1
end
return 0
`

	released, err := h.redisClient.Eval(ctx, releaseScript, []string{onlineKey, lastSeenKey}, sessionID, nowText, strconv.FormatInt(int64(service.LastSeenTTL/time.Second), 10)).Int64()
	if err != nil {
		return false, ""
	}
	return released == 1, nowText
}

func (h *WSHandler) broadcastOnlineStatus(userID uint64, online bool, lastSeenAt string) {
	notification := ws.OutgoingEnvelope{
		Type: "user_online_status",
		Payload: ws.OutgoingOnlineStatusPayload{
			UserID:     userID,
			Online:     online,
			LastSeenAt: lastSeenAt,
		},
	}
	payload, err := json.Marshal(notification)
	if err != nil {
		return
	}
	h.manager.BroadcastAllLocal(payload)
}
