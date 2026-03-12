package service

import (
	"context"
	"errors"
	"strings"

	"go-chat/internal/model"
	"go-chat/internal/repository"
)

type ChatService struct {
	messages repository.MessageRepository
	groups   repository.GroupRepository
	users    repository.UserRepository
	friends  repository.FriendRepository
}

func NewChatService(messages repository.MessageRepository, groups repository.GroupRepository, users repository.UserRepository, friends repository.FriendRepository) *ChatService {
	return &ChatService{
		messages: messages,
		groups:   groups,
		users:    users,
		friends:  friends,
	}
}

func (s *ChatService) SendPrivateMessage(ctx context.Context, senderID, targetID uint64, msgType int8, content string) (*model.Message, []uint64, error) {
	normalizedMsgType, normalizedContent, err := normalizeMessagePayload(msgType, content)
	if err != nil {
		return nil, nil, err
	}

	// 检查目标用户是否存在
	target, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		return nil, nil, err
	}
	if target == nil {
		return nil, nil, ErrTargetUserNotFound
	}

	// 检查好友关系
	isFriend, err := s.friends.IsFriend(ctx, senderID, targetID)
	if err != nil {
		return nil, nil, err
	}
	if !isFriend {
		return nil, nil, ErrNotFriend
	}

	message := &model.Message{
		SenderID:   senderID,
		ReceiverID: &targetID,
		Content:    normalizedContent,
		MsgType:    normalizedMsgType,
	}

	if err := s.messages.Create(ctx, message); err != nil {
		return nil, nil, err
	}

	return message, uniqueUserIDs(senderID, targetID), nil
}

func (s *ChatService) SendGroupMessage(ctx context.Context, senderID, groupID uint64, msgType int8, content string) (*model.Message, []uint64, error) {
	normalizedMsgType, normalizedContent, err := normalizeMessagePayload(msgType, content)
	if err != nil {
		return nil, nil, err
	}

	// 检查群组是否存在
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	if group == nil {
		return nil, nil, ErrGroupNotFound
	}

	// 检查用户是否属于群组
	isMember, err := s.groups.IsMember(ctx, groupID, senderID)
	if err != nil {
		return nil, nil, err
	}
	if !isMember {
		return nil, nil, ErrNotGroupMember
	}

	message := &model.Message{
		SenderID: senderID,
		GroupID:  &groupID,
		Content:  normalizedContent,
		MsgType:  normalizedMsgType,
	}

	if err := s.messages.Create(ctx, message); err != nil {
		return nil, nil, err
	}

	// 获取群组成员ID列表
	memberIDs, err := s.groups.ListMemberIDs(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	if len(memberIDs) == 0 {
		memberIDs = []uint64{senderID}
	}

	return message, uniqueUserIDs(memberIDs...), nil
}

func (s *ChatService) ListHistory(ctx context.Context, userID, targetID uint64, chatType string, lastMsgID, aroundMsgID uint64, limit int, keyword string) (*model.MessageHistoryPage, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	normalizedKeyword := strings.TrimSpace(keyword)
	if len(normalizedKeyword) > 100 {
		return nil, ErrInvalidKeyword
	}
	if aroundMsgID > 0 && (lastMsgID > 0 || normalizedKeyword != "") {
		return nil, ErrInvalidParamCombo
	}

	switch chatType {
	case "private":
		target, err := s.users.GetByID(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, ErrTargetUserNotFound
		}
		//查询数据库私聊历史
		var (
			messages []model.Message
			hasMore  bool
		)
		if aroundMsgID > 0 {
			messages, hasMore, err = s.messages.ListPrivateHistoryAround(ctx, userID, targetID, aroundMsgID, limit)
		} else {
			messages, hasMore, err = s.messages.ListPrivateHistory(ctx, userID, targetID, lastMsgID, limit, normalizedKeyword)
		}
		if err != nil {
			if errors.Is(err, repository.ErrInvalidHistoryCursor) {
				return nil, ErrInvalidLastMsgID
			}
			if errors.Is(err, repository.ErrInvalidHistoryAnchor) {
				return nil, ErrInvalidAroundMsgID
			}
			return nil, err
		}
		if err := s.attachSenderProfiles(ctx, messages); err != nil {
			return nil, err
		}
		return buildHistoryPage(messages, hasMore), nil
	case "group":
		group, err := s.groups.GetByID(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if group == nil {
			return nil, ErrGroupNotFound
		}

		//检查用户是否属于群组
		isMember, err := s.groups.IsMember(ctx, targetID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrNotGroupMember
		}

		var (
			messages []model.Message
			hasMore  bool
		)
		if aroundMsgID > 0 {
			messages, hasMore, err = s.messages.ListGroupHistoryAround(ctx, targetID, aroundMsgID, limit)
		} else {
			messages, hasMore, err = s.messages.ListGroupHistory(ctx, targetID, lastMsgID, limit, normalizedKeyword)
		}
		if err != nil {
			if errors.Is(err, repository.ErrInvalidHistoryCursor) {
				return nil, ErrInvalidLastMsgID
			}
			if errors.Is(err, repository.ErrInvalidHistoryAnchor) {
				return nil, ErrInvalidAroundMsgID
			}
			return nil, err
		}
		if err := s.attachSenderProfiles(ctx, messages); err != nil {
			return nil, err
		}
		return buildHistoryPage(messages, hasMore), nil
	default:
		return nil, ErrInvalidChatType
	}
}

func buildHistoryPage(messages []model.Message, hasMore bool) *model.MessageHistoryPage {
	page := &model.MessageHistoryPage{
		Items:   messages,
		HasMore: hasMore,
	}
	if hasMore && len(messages) > 0 {
		nextCursor := messages[0].ID
		page.NextCursor = &nextCursor
	}
	return page
}

func (s *ChatService) ListPrivateConversations(ctx context.Context, userID uint64, limit int) ([]model.PrivateConversationSummary, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	return s.messages.ListPrivateConversations(ctx, userID, limit)
}

func (s *ChatService) GetConversationReadInfo(ctx context.Context, userID, targetID uint64, chatType string) (*model.ConversationReadInfo, error) {
	normalizedChatType := strings.ToLower(strings.TrimSpace(chatType))
	info := &model.ConversationReadInfo{
		ChatType: normalizedChatType,
		TargetID: targetID,
	}

	switch normalizedChatType {
	case "private":
		target, err := s.users.GetByID(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, ErrTargetUserNotFound
		}

		state, err := s.messages.GetReadState(ctx, userID, normalizedChatType, targetID)
		if err != nil {
			return nil, err
		}
		if state != nil {
			info.LastReadMessageID = state.LastReadMessageID
		}

		unreadCount, err := s.messages.CountUnreadPrivate(ctx, userID, targetID, info.LastReadMessageID)
		if err != nil {
			return nil, err
		}
		info.UnreadCount = unreadCount
		return info, nil

	case "group":
		group, err := s.groups.GetByID(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if group == nil {
			return nil, ErrGroupNotFound
		}

		isMember, err := s.groups.IsMember(ctx, targetID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrNotGroupMember
		}

		state, err := s.messages.GetReadState(ctx, userID, normalizedChatType, targetID)
		if err != nil {
			return nil, err
		}
		if state != nil {
			info.LastReadMessageID = state.LastReadMessageID
		}

		unreadCount, err := s.messages.CountUnreadGroup(ctx, userID, targetID, info.LastReadMessageID)
		if err != nil {
			return nil, err
		}
		info.UnreadCount = unreadCount
		return info, nil
	default:
		return nil, ErrInvalidChatType
	}
}

func (s *ChatService) MarkConversationRead(ctx context.Context, userID, targetID uint64, chatType string) (*model.ConversationReadInfo, error) {
	normalizedChatType := strings.ToLower(strings.TrimSpace(chatType))

	// 先做目标合法性检查
	if _, err := s.GetConversationReadInfo(ctx, userID, targetID, normalizedChatType); err != nil {
		return nil, err
	}

	var (
		latestMessageID uint64
		err             error
	)

	switch normalizedChatType {
	case "private":
		latestMessageID, err = s.messages.GetLatestPrivateMessageID(ctx, userID, targetID)
	case "group":
		latestMessageID, err = s.messages.GetLatestGroupMessageID(ctx, targetID)
	default:
		return nil, ErrInvalidChatType
	}
	if err != nil {
		return nil, err
	}

	if err := s.messages.UpsertReadState(ctx, userID, normalizedChatType, targetID, latestMessageID); err != nil {
		return nil, err
	}

	info, err := s.GetConversationReadInfo(ctx, userID, targetID, normalizedChatType)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// uniqueUserIDs 是一个简单的工具函数，用来给 ID 列表去重
func uniqueUserIDs(ids ...uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(ids))
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func normalizeMessagePayload(msgType int8, content string) (int8, string, error) {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return 0, "", ErrInvalidContent
	}

	if msgType == 0 {
		msgType = 1
	}
	if msgType != 1 && msgType != 2 {
		return 0, "", ErrInvalidMsgType
	}

	return msgType, trimmedContent, nil
}

func (s *ChatService) attachSenderProfiles(ctx context.Context, messages []model.Message) error {
	if len(messages) == 0 {
		return nil
	}

	uniqueIDs := make(map[uint64]struct{}, len(messages))
	ids := make([]uint64, 0, len(messages))
	for _, message := range messages {
		if _, exists := uniqueIDs[message.SenderID]; exists {
			continue
		}
		uniqueIDs[message.SenderID] = struct{}{}
		ids = append(ids, message.SenderID)
	}

	users, err := s.users.ListByIDs(ctx, ids)
	if err != nil {
		return err
	}

	userMap := make(map[uint64]model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	for i := range messages {
		user, ok := userMap[messages[i].SenderID]
		if !ok {
			continue
		}
		messages[i].Sender = &model.SenderProfile{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
		}
	}

	return nil
}
