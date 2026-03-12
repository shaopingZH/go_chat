package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go-chat/internal/api/middleware"
	"go-chat/internal/service"
)

type MessageHandler struct {
	chatService *service.ChatService
}

type markConversationReadBody struct {
	ChatType string `json:"chat_type"`
	TargetID uint64 `json:"target_id"`
}

func NewMessageHandler(chatService *service.ChatService) *MessageHandler {
	return &MessageHandler{chatService: chatService}
}

func (h *MessageHandler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// c.Query("target_id") 获取网址问号后面的参数
	// strconv.ParseUint 把字符串转成 64 位无符号整数
	targetID, err := strconv.ParseUint(c.Query("target_id"), 10, 64)
	if err != nil || targetID == 0 {
		writeError(c, http.StatusBadRequest, "invalid target_id")
		return
	}

	chatType := c.DefaultQuery("type", "private")

	var lastMsgID uint64
	lastMsgRaw := c.Query("last_msg_id") // 拿上一条消息的 ID
	if lastMsgRaw != "" {
		// 为什么要传这个？因为翻页时，我们要告诉数据库：“给我 ID 比 1000 更早的消息”
		lastMsgID, err = strconv.ParseUint(lastMsgRaw, 10, 64) //字符串转数字
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid last_msg_id")
			return
		}
	}

	var aroundMsgID uint64
	aroundMsgRaw := c.Query("around_msg_id")
	if aroundMsgRaw != "" {
		aroundMsgID, err = strconv.ParseUint(aroundMsgRaw, 10, 64)
		if err != nil || aroundMsgID == 0 {
			writeError(c, http.StatusBadRequest, "invalid around_msg_id")
			return
		}
	}

	limit := 30
	limitRaw := c.Query("limit")
	if limitRaw != "" {
		// 把字符串转成普通的整型 (int)
		limit, err = strconv.Atoi(limitRaw)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid limit")
			return
		}
	}

	keyword := ""
	if keywordRaw, exists := c.GetQuery("keyword"); exists {
		keyword = strings.TrimSpace(keywordRaw)
		if keyword == "" {
			writeError(c, http.StatusBadRequest, "invalid keyword")
			return
		}
	}

	if aroundMsgID > 0 && (lastMsgRaw != "" || keyword != "") {
		writeError(c, http.StatusBadRequest, service.ErrInvalidParamCombo.Error())
		return
	}

	historyPage, err := h.chatService.ListHistory(c.Request.Context(), userID, targetID, chatType, lastMsgID, aroundMsgID, limit, keyword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTargetUserNotFound), errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidChatType), errors.Is(err, service.ErrNotGroupMember), errors.Is(err, service.ErrInvalidLastMsgID), errors.Is(err, service.ErrInvalidAroundMsgID), errors.Is(err, service.ErrInvalidKeyword), errors.Is(err, service.ErrInvalidParamCombo):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	readInfo, err := h.chatService.GetConversationReadInfo(c.Request.Context(), userID, targetID, chatType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTargetUserNotFound), errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidChatType), errors.Is(err, service.ErrNotGroupMember):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":                historyPage.Items,
		"last_read_message_id": readInfo.LastReadMessageID,
		"unread_count":         readInfo.UnreadCount,
		"has_more":             historyPage.HasMore,
		"next_cursor":          historyPage.NextCursor,
	})
}

func (h *MessageHandler) ListPrivateConversations(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := 30
	limitRaw := c.Query("limit")
	if limitRaw != "" {
		value, err := strconv.Atoi(limitRaw)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = value
	}

	items, err := h.chatService.ListPrivateConversations(c.Request.Context(), userID, limit)
	if err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *MessageHandler) MarkConversationRead(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body markConversationReadBody
	if err := c.ShouldBindJSON(&body); err != nil || body.TargetID == 0 {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	info, err := h.chatService.MarkConversationRead(c.Request.Context(), userID, body.TargetID, body.ChatType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTargetUserNotFound), errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidChatType), errors.Is(err, service.ErrNotGroupMember):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, info)
}
