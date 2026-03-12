package repository

import (
	"context"
	"errors"

	"go-chat/internal/model"
	"gorm.io/gorm"
)

type FriendRepository interface {
	// 好友请求
	CreateRequest(ctx context.Context, request *model.FriendRequest) error
	GetRequestByID(ctx context.Context, requestID uint64) (*model.FriendRequest, error)
	GetPendingRequest(ctx context.Context, senderID, receiverID uint64) (*model.FriendRequest, error)
	UpdateRequestStatus(ctx context.Context, requestID uint64, status string) error
	ListPendingByReceiverID(ctx context.Context, receiverID uint64) ([]model.FriendRequest, error)

	// 好友关系
	IsFriend(ctx context.Context, userID, friendID uint64) (bool, error)
	CreateFriendship(ctx context.Context, userID, friendID uint64) error
	DeleteFriendship(ctx context.Context, userID, friendID uint64) error
	ListFriendsWithProfile(ctx context.Context, userID uint64) ([]model.User, error)
}

type friendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{db: db}
}

// ---------- 好友请求 ----------

func (r *friendRepository) CreateRequest(ctx context.Context, request *model.FriendRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *friendRepository) GetRequestByID(ctx context.Context, requestID uint64) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := r.db.WithContext(ctx).First(&req, requestID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *friendRepository) GetPendingRequest(ctx context.Context, senderID, receiverID uint64) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("sender_id = ? AND receiver_id = ? AND status = 'pending'", senderID, receiverID).
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *friendRepository) UpdateRequestStatus(ctx context.Context, requestID uint64, status string) error {
	result := r.db.WithContext(ctx).
		Model(&model.FriendRequest{}).
		Where("id = ? AND status = 'pending'", requestID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *friendRepository) ListPendingByReceiverID(ctx context.Context, receiverID uint64) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	err := r.db.WithContext(ctx).
		Where("receiver_id = ? AND status = 'pending'", receiverID).
		Order("created_at DESC").
		Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// ---------- 好友关系 ----------

func (r *friendRepository) IsFriend(ctx context.Context, userID, friendID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *friendRepository) CreateFriendship(ctx context.Context, userID, friendID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pair := []model.Friendship{
			{UserID: userID, FriendID: friendID},
			{UserID: friendID, FriendID: userID},
		}
		for _, f := range pair {
			if err := tx.Create(&f).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *friendRepository) DeleteFriendship(ctx context.Context, userID, friendID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).
			Delete(&model.Friendship{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND friend_id = ?", friendID, userID).
			Delete(&model.Friendship{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *friendRepository) ListFriendsWithProfile(ctx context.Context, userID uint64) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Table("users u").
		Select("u.id, u.username, u.display_name, u.avatar, u.bio").
		Joins("JOIN friendships f ON f.friend_id = u.id").
		Where("f.user_id = ?", userID).
		Order("f.created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
