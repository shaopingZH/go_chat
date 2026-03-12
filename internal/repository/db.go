package repository

import (
	"time"

	"go-chat/internal/model"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQL(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.GroupMember{},
		&model.Message{},
		&model.FriendRequest{},
		&model.Friendship{},
		&model.ConversationReadState{},
	)
}
