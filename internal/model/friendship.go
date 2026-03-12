package model

import "time"

type Friendship struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"uniqueIndex:uk_user_friend;not null" json:"user_id"`
	FriendID  uint64    `gorm:"uniqueIndex:uk_user_friend;not null" json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`
}
