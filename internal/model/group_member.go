package model

import "time"

type GroupMember struct {
	ID       uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID  uint64    `gorm:"uniqueIndex:uk_group_user;not null" json:"group_id"`
	UserID   uint64    `gorm:"uniqueIndex:uk_group_user;not null" json:"user_id"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}
