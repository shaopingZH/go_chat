package model

import "time"

type PrivateConversationSummary struct {
	PartnerID       uint64     `json:"partner_id"`
	PartnerUsername string     `json:"partner_username"`
	PartnerAvatar   string     `json:"partner_avatar"`
	UnreadCount     int64      `json:"unread_count"`
	LastMessageID   *uint64    `json:"last_message_id,omitempty"`
	LastSenderID    *uint64    `json:"last_sender_id,omitempty"`
	LastMsgType     *int8      `json:"last_msg_type,omitempty"`
	LastContent     string     `json:"last_content"`
	LastCreatedAt   *time.Time `json:"last_created_at,omitempty"`
}

type GroupLastMessageSummary struct {
	ID        uint64    `json:"id"`
	SenderID  uint64    `json:"sender_id"`
	MsgType   int8      `json:"msg_type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupConversationSummary struct {
	ID          uint64                   `json:"id"`
	Name        string                   `json:"name"`
	Avatar      string                   `json:"avatar"`
	OwnerID     uint64                   `json:"owner_id"`
	CreatedAt   time.Time                `json:"created_at"`
	UnreadCount int64                    `json:"unread_count"`
	LastMessage *GroupLastMessageSummary `json:"last_message,omitempty"`
}

type ConversationReadInfo struct {
	ChatType          string `json:"chat_type"`
	TargetID          uint64 `json:"target_id"`
	LastReadMessageID uint64 `json:"last_read_message_id"`
	UnreadCount       int64  `json:"unread_count"`
}
