package service

import (
	"context"

	"go-chat/internal/model"
	"go-chat/internal/repository"
)

type FriendService struct {
	friends repository.FriendRepository
	users   repository.UserRepository
}

func NewFriendService(friends repository.FriendRepository, users repository.UserRepository) *FriendService {
	return &FriendService{friends: friends, users: users}
}

// SendRequest 发送好友请求
func (s *FriendService) SendRequest(ctx context.Context, senderID, targetID uint64) (*model.FriendRequest, error) {
	if senderID == targetID {
		return nil, ErrCannotAddSelf
	}

	// 检查目标用户存在
	target, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, ErrTargetUserNotFound
	}

	// 检查是否已经是好友
	isFriend, err := s.friends.IsFriend(ctx, senderID, targetID)
	if err != nil {
		return nil, err
	}
	if isFriend {
		return nil, ErrAlreadyFriends
	}

	// 检查是否已有 pending 请求（双向都检查）
	pending, err := s.friends.GetPendingRequest(ctx, senderID, targetID)
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return nil, ErrRequestAlreadyPending
	}

	// 也检查对方是否已经向我发送了 pending 请求
	reversePending, err := s.friends.GetPendingRequest(ctx, targetID, senderID)
	if err != nil {
		return nil, err
	}
	if reversePending != nil {
		return nil, ErrRequestAlreadyPending
	}

	req := &model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: targetID,
		Status:     "pending",
	}
	if err := s.friends.CreateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

// AcceptRequest 接受好友请求，返回请求记录（含 SenderID 供通知）
func (s *FriendService) AcceptRequest(ctx context.Context, requestID, currentUserID uint64) (*model.FriendRequest, error) {
	req, err := s.friends.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, ErrRequestNotFound
	}
	if req.ReceiverID != currentUserID {
		return nil, ErrNotYourRequest
	}
	if req.Status != "pending" {
		return nil, ErrRequestAlreadyProcessed
	}

	// 更新请求状态
	if err := s.friends.UpdateRequestStatus(ctx, requestID, "accepted"); err != nil {
		return nil, err
	}

	// 创建双向好友关系
	if err := s.friends.CreateFriendship(ctx, req.SenderID, req.ReceiverID); err != nil {
		return nil, err
	}

	return req, nil
}

// RejectRequest 拒绝好友请求
func (s *FriendService) RejectRequest(ctx context.Context, requestID, currentUserID uint64) error {
	req, err := s.friends.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req == nil {
		return ErrRequestNotFound
	}
	if req.ReceiverID != currentUserID {
		return ErrNotYourRequest
	}
	if req.Status != "pending" {
		return ErrRequestAlreadyProcessed
	}

	return s.friends.UpdateRequestStatus(ctx, requestID, "rejected")
}

// ListFriends 获取好友列表（含公开资料）
func (s *FriendService) ListFriends(ctx context.Context, userID uint64) ([]model.User, error) {
	return s.friends.ListFriendsWithProfile(ctx, userID)
}

// ListPendingRequests 获取收到的待处理请求
func (s *FriendService) ListPendingRequests(ctx context.Context, userID uint64) ([]model.FriendRequest, error) {
	return s.friends.ListPendingByReceiverID(ctx, userID)
}

// DeleteFriend 删除好友（双向）
func (s *FriendService) DeleteFriend(ctx context.Context, userID, friendID uint64) error {
	isFriend, err := s.friends.IsFriend(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !isFriend {
		return ErrFriendshipNotFound
	}

	return s.friends.DeleteFriendship(ctx, userID, friendID)
}

// IsFriend 检查两个用户是否为好友（供 ChatService 调用）
func (s *FriendService) IsFriend(ctx context.Context, userID, friendID uint64) (bool, error) {
	return s.friends.IsFriend(ctx, userID, friendID)
}
