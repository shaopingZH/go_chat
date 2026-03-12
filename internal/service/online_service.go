package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go-chat/internal/repository"
)

const (
	OnlineTTL   = 5 * time.Minute
	LastSeenTTL = 30 * 24 * time.Hour
)

type LocalOnlineChecker interface {
	IsOnline(userID uint64) bool
}

type LocalLastSeenProvider interface {
	LastSeen(userID uint64) (time.Time, bool)
}

type UserOnlineStatus struct {
	UserID     uint64 `json:"user_id"`
	Online     bool   `json:"online"`
	LastSeenAt string `json:"last_seen_at,omitempty"`
}

type OnlineService struct {
	users        repository.UserRepository
	redisClient  *redis.Client
	localChecker LocalOnlineChecker
}

func NewOnlineService(users repository.UserRepository, redisClient *redis.Client, localChecker LocalOnlineChecker) *OnlineService {
	return &OnlineService{users: users, redisClient: redisClient, localChecker: localChecker}
}

func UserOnlineKey(userID uint64) string {
	return fmt.Sprintf("user:online:%d", userID)
}

func UserLastSeenKey(userID uint64) string {
	return fmt.Sprintf("user:last_seen:%d", userID)
}

func (s *OnlineService) GetUserOnlineStatus(ctx context.Context, userID uint64) (*UserOnlineStatus, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	status, err := s.buildStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (s *OnlineService) BatchGetUserOnlineStatus(ctx context.Context, userIDs []uint64) ([]UserOnlineStatus, error) {
	if len(userIDs) == 0 {
		return nil, ErrInvalidUserIDs
	}

	orderedUnique := make([]uint64, 0, len(userIDs))
	seen := make(map[uint64]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == 0 {
			return nil, ErrInvalidUserIDs
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		orderedUnique = append(orderedUnique, userID)
	}

	users, err := s.users.ListByIDs(ctx, orderedUnique)
	if err != nil {
		return nil, err
	}

	exists := make(map[uint64]struct{}, len(users))
	for _, user := range users {
		exists[user.ID] = struct{}{}
	}

	items := make([]UserOnlineStatus, 0, len(users))
	for _, userID := range orderedUnique {
		if _, ok := exists[userID]; !ok {
			continue
		}
		status, buildErr := s.buildStatus(ctx, userID)
		if buildErr != nil {
			return nil, buildErr
		}
		items = append(items, *status)
	}

	return items, nil
}

func (s *OnlineService) buildStatus(ctx context.Context, userID uint64) (*UserOnlineStatus, error) {
	if s.localChecker != nil && s.localChecker.IsOnline(userID) {
		return &UserOnlineStatus{UserID: userID, Online: true}, nil
	}

	if s.redisClient != nil {
		onlineKey := UserOnlineKey(userID)
		if _, err := s.redisClient.Get(ctx, onlineKey).Result(); err == nil {
			return &UserOnlineStatus{UserID: userID, Online: true}, nil
		} else if err != redis.Nil {
			return nil, err
		}

		lastSeenKey := UserLastSeenKey(userID)
		lastSeen, err := s.redisClient.Get(ctx, lastSeenKey).Result()
		if err == nil {
			return &UserOnlineStatus{UserID: userID, Online: false, LastSeenAt: lastSeen}, nil
		}
		if err != redis.Nil {
			return nil, err
		}
	}

	if provider, ok := s.localChecker.(LocalLastSeenProvider); ok {
		if value, exists := provider.LastSeen(userID); exists {
			return &UserOnlineStatus{UserID: userID, Online: false, LastSeenAt: value.UTC().Format(time.RFC3339)}, nil
		}
	}

	return &UserOnlineStatus{UserID: userID, Online: false}, nil
}
