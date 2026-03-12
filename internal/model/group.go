package model

import "time"

type Group struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Avatar    string    `gorm:"type:varchar(255);not null;default:''" json:"avatar"`
	OwnerID   uint64    `gorm:"index;not null" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Group) TableName() string {
	return "groups"
}
