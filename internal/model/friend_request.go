package model

import "time"

type FriendRequest struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID   uint64    `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint64    `gorm:"not null;index" json:"receiver_id"`
	Status     string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
