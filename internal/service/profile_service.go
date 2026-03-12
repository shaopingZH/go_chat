package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"go-chat/internal/model"
	"go-chat/internal/repository"
)

const (
	maxDisplayNameLen = 50
	maxBioLen         = 200
	avatarPrefix      = "/uploads/images/"
)

type UpdateProfileInput struct {
	DisplayName *string
	Avatar      *string
	Bio         *string
}

type ProfileService struct {
	users repository.UserRepository
}

func NewProfileService(users repository.UserRepository) *ProfileService {
	return &ProfileService{users: users}
}

func (s *ProfileService) GetMyProfile(ctx context.Context, userID uint64) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *ProfileService) GetPublicProfile(ctx context.Context, userID uint64) (*model.User, error) {
	return s.GetMyProfile(ctx, userID)
}

func (s *ProfileService) UpdateMyProfile(ctx context.Context, userID uint64, input UpdateProfileInput) (*model.User, error) {
	if input.DisplayName == nil && input.Avatar == nil && input.Bio == nil {
		return nil, ErrInvalidProfilePayload
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	displayName := user.DisplayName
	avatar := user.Avatar
	bio := user.Bio

	if input.DisplayName != nil {
		next := strings.TrimSpace(*input.DisplayName)
		if utf8.RuneCountInString(next) > maxDisplayNameLen {
			return nil, ErrInvalidDisplayName
		}
		displayName = next
	}

	if input.Avatar != nil {
		next := strings.TrimSpace(*input.Avatar)
		if next != "" && !strings.HasPrefix(next, avatarPrefix) {
			return nil, ErrInvalidAvatar
		}
		avatar = next
	}

	if input.Bio != nil {
		next := strings.TrimSpace(*input.Bio)
		if utf8.RuneCountInString(next) > maxBioLen {
			return nil, ErrInvalidBio
		}
		bio = next
	}

	if err := s.users.UpdateProfileByID(ctx, userID, displayName, avatar, bio); err != nil {
		return nil, err
	}

	updated, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrUserNotFound
	}

	return updated, nil
}
