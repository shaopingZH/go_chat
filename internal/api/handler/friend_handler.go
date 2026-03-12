package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-chat/internal/api/middleware"
	"go-chat/internal/repository"
	"go-chat/internal/service"
	"go-chat/internal/ws"
)

type FriendHandler struct {
	friendService *service.FriendService
	userRepo      repository.UserRepository
	wsManager     *ws.Manager
}

func NewFriendHandler(friendService *service.FriendService, userRepo repository.UserRepository, wsManager *ws.Manager) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
		userRepo:      userRepo,
		wsManager:     wsManager,
	}
}

type sendFriendRequestBody struct {
	TargetID uint64 `json:"target_id"`
}

// SearchUsers 搜索用户（按用户名模糊匹配）
func (h *FriendHandler) SearchUsers(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		writeError(c, http.StatusBadRequest, "invalid keyword")
		return
	}

	users, err := h.userRepo.SearchByUsername(c.Request.Context(), keyword, userID, 20)
	if err != nil {
		writeInternalError(c)
		return
	}

	items := make([]gin.H, 0, len(users))
	for _, u := range users {
		items = append(items, gin.H{
			"id":           u.ID,
			"username":     u.Username,
			"display_name": u.DisplayName,
			"avatar":       u.Avatar,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// SendRequest 发送好友请求
func (h *FriendHandler) SendRequest(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body sendFriendRequestBody
	if err := c.ShouldBindJSON(&body); err != nil || body.TargetID == 0 {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	req, err := h.friendService.SendRequest(c.Request.Context(), userID, body.TargetID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotAddSelf):
			writeError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrTargetUserNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrRequestAlreadyPending):
			writeError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrAlreadyFriends):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	// WebSocket 通知目标用户
	sender, _ := h.userRepo.GetByID(c.Request.Context(), userID)
	if sender != nil {
		notification := ws.OutgoingEnvelope{
			Type: "friend_request",
			Payload: gin.H{
				"id":         req.ID,
				"request_id": req.ID,
				"status":     req.Status,
				"from_user": gin.H{
					"id":           sender.ID,
					"username":     sender.Username,
					"display_name": sender.DisplayName,
					"avatar":       sender.Avatar,
				},
				"created_at": req.CreatedAt.UTC().Format(time.RFC3339),
			},
		}
		if payload, err := json.Marshal(notification); err == nil {
			h.wsManager.SendToUser(body.TargetID, payload)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          req.ID,
		"sender_id":   req.SenderID,
		"receiver_id": req.ReceiverID,
		"status":      req.Status,
		"created_at":  req.CreatedAt.UTC().Format(time.RFC3339),
	})
}

// AcceptRequest 接受好友请求
func (h *FriendHandler) AcceptRequest(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || requestID == 0 {
		writeError(c, http.StatusBadRequest, "invalid request id")
		return
	}

	req, err := h.friendService.AcceptRequest(c.Request.Context(), requestID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequestNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotYourRequest):
			writeError(c, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrRequestAlreadyProcessed):
			writeError(c, http.StatusConflict, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	// WebSocket 通知发送方
	acceptor, _ := h.userRepo.GetByID(c.Request.Context(), userID)
	if acceptor != nil {
		notification := ws.OutgoingEnvelope{
			Type: "friend_accepted",
			Payload: gin.H{
				"request_id": req.ID,
				"friend": gin.H{
					"id":           acceptor.ID,
					"username":     acceptor.Username,
					"display_name": acceptor.DisplayName,
					"avatar":       acceptor.Avatar,
				},
				"accepted_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		if payload, err := json.Marshal(notification); err == nil {
			h.wsManager.SendToUser(req.SenderID, payload)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "accepted"})
}

// RejectRequest 拒绝好友请求
func (h *FriendHandler) RejectRequest(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || requestID == 0 {
		writeError(c, http.StatusBadRequest, "invalid request id")
		return
	}

	err = h.friendService.RejectRequest(c.Request.Context(), requestID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRequestNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotYourRequest):
			writeError(c, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrRequestAlreadyProcessed):
			writeError(c, http.StatusConflict, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rejected"})
}

// ListFriends 好友列表
func (h *FriendHandler) ListFriends(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	friends, err := h.friendService.ListFriends(c.Request.Context(), userID)
	if err != nil {
		writeInternalError(c)
		return
	}

	items := make([]gin.H, 0, len(friends))
	for _, f := range friends {
		items = append(items, gin.H{
			"id":           f.ID,
			"username":     f.Username,
			"display_name": f.DisplayName,
			"avatar":       f.Avatar,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ListPendingRequests 待处理请求列表
func (h *FriendHandler) ListPendingRequests(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	requests, err := h.friendService.ListPendingRequests(c.Request.Context(), userID)
	if err != nil {
		writeInternalError(c)
		return
	}

	items := make([]gin.H, 0, len(requests))
	for _, r := range requests {
		fromUser := gin.H{
			"id":           r.SenderID,
			"username":     "",
			"display_name": "",
			"avatar":       "",
		}

		// 查询发送者信息
		sender, err := h.userRepo.GetByID(c.Request.Context(), r.SenderID)
		if err == nil && sender != nil {
			fromUser = gin.H{
				"id":           sender.ID,
				"username":     sender.Username,
				"display_name": sender.DisplayName,
				"avatar":       sender.Avatar,
			}
		}

		items = append(items, gin.H{
			"request_id": r.ID,
			"from_user":  fromUser,
			"created_at": r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// DeleteFriend 删除好友
func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	friendID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || friendID == 0 {
		writeError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.friendService.DeleteFriend(c.Request.Context(), userID, friendID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFriendshipNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
