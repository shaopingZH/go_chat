package model

import "time"

type ConversationReadState struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint64    `gorm:"not null;uniqueIndex:uk_user_chat_target;index" json:"user_id"`
	ChatType          string    `gorm:"type:varchar(16);not null;uniqueIndex:uk_user_chat_target" json:"chat_type"`
	TargetID          uint64    `gorm:"not null;uniqueIndex:uk_user_chat_target;index" json:"target_id"`
	LastReadMessageID uint64    `gorm:"not null;default:0" json:"last_read_message_id"`
	LastReadAt        time.Time `gorm:"not null" json:"last_read_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
