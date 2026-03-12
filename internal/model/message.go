package model

import "time"

type Message struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID   uint64         `gorm:"index;not null" json:"sender_id"`
	ReceiverID *uint64        `gorm:"index" json:"receiver_id,omitempty"`
	GroupID    *uint64        `gorm:"index" json:"group_id,omitempty"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	MsgType    int8           `gorm:"default:1" json:"msg_type"`
	CreatedAt  time.Time      `json:"created_at"`
	Sender     *SenderProfile `gorm:"-" json:"sender,omitempty"`
}

type SenderProfile struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
}

type MessageHistoryPage struct {
	Items      []Message `json:"items"`
	HasMore    bool      `json:"has_more"`
	NextCursor *uint64   `json:"next_cursor,omitempty"`
}
