package model

import "time"

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	DisplayName  string    `gorm:"type:varchar(50);default:''" json:"display_name"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Avatar       string    `gorm:"type:varchar(255);default:''" json:"avatar"`
	Bio          string    `gorm:"type:varchar(200);default:''" json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
