package service

import (
	"context"
	"strings"
	"time"

	"go-chat/internal/config"
	"go-chat/internal/model"
	"go-chat/internal/repository"
	"go-chat/pkg/jwtutil"
	"go-chat/pkg/password"
)

type AuthService struct {
	users repository.UserRepository
	cfg   *config.Config
}

func NewAuthService(users repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, username, rawPassword string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || len(username) > 50 || len(rawPassword) < 6 {
		return nil, ErrInvalidRegister
	}

	//查用户名在数据库是否存在
	existing, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	//密码加密
	passwordHash, err := password.Hash(rawPassword)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: passwordHash,
		Avatar:       "",
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username, rawPassword string) (string, *model.User, error) {
	//查找用户
	user, err := s.users.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return "", nil, err
	}
	//验证密码
	if user == nil || !password.Verify(user.PasswordHash, rawPassword) {
		return "", nil, ErrInvalidCredentials
	}

	//发放token
	token, err := jwtutil.GenerateToken(
		s.cfg.JWTSecret,
		user.ID,
		user.Username,
		time.Duration(s.cfg.JWTExpireHours)*time.Hour,
	)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uint64) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *AuthService) ParseToken(rawToken string) (*jwtutil.Claims, error) {
	return jwtutil.ParseToken(s.cfg.JWTSecret, rawToken)
}
