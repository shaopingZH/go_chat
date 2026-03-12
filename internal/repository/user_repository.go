package repository

import (
	"context"
	"errors"

	"go-chat/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, userID uint64) (*model.User, error)
	ListByIDs(ctx context.Context, userIDs []uint64) ([]model.User, error)
	UpdateProfileByID(ctx context.Context, userID uint64, displayName, avatar, bio string) error
	SearchByUsername(ctx context.Context, keyword string, excludeUserID uint64, limit int) ([]model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) ListByIDs(ctx context.Context, userIDs []uint64) ([]model.User, error) {
	if len(userIDs) == 0 {
		return []model.User{}, nil
	}

	var users []model.User
	err := r.db.WithContext(ctx).
		Select("id, username, display_name, avatar").
		Where("id IN ?", userIDs).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) UpdateProfileByID(ctx context.Context, userID uint64, displayName, avatar, bio string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"display_name": displayName,
			"avatar":       avatar,
			"bio":          bio,
		}).Error
}

func (r *userRepository) SearchByUsername(ctx context.Context, keyword string, excludeUserID uint64, limit int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Select("id, username, display_name, avatar").
		Where("username LIKE ? AND id != ?", "%"+keyword+"%", excludeUserID).
		Order("username ASC").
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
