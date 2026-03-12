package service

import (
	"context"
	"strings"

	"go-chat/internal/model"
	"go-chat/internal/repository"
)

type GroupService struct {
	groups repository.GroupRepository
}

func NewGroupService(groups repository.GroupRepository) *GroupService {
	return &GroupService{groups: groups}
}

func (s *GroupService) CreateGroup(ctx context.Context, ownerID uint64, name string) (*model.Group, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 100 {
		return nil, ErrInvalidGroupName
	}

	group := &model.Group{
		Name:    name,
		OwnerID: ownerID,
	}

	if err := s.groups.Create(ctx, group); err != nil {
		return nil, err
	}

	if err := s.groups.AddMember(ctx, group.ID, ownerID); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *GroupService) UpdateGroupProfile(ctx context.Context, currentUserID, groupID uint64, name, avatar *string) (*model.Group, error) {
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}
	if group.OwnerID != currentUserID {
		return nil, ErrNotGroupOwner
	}

	nextName := group.Name
	if name != nil {
		nextName = strings.TrimSpace(*name)
		if nextName == "" || len(nextName) > 100 {
			return nil, ErrInvalidGroupName
		}
	}

	nextAvatar := group.Avatar
	if avatar != nil {
		nextAvatar = strings.TrimSpace(*avatar)
		if nextAvatar != "" && !strings.HasPrefix(nextAvatar, avatarPrefix) {
			return nil, ErrInvalidAvatar
		}
	}

	if err := s.groups.UpdateProfileByID(ctx, groupID, nextName, nextAvatar); err != nil {
		return nil, err
	}
	group.Name = nextName
	group.Avatar = nextAvatar
	return group, nil
}

func (s *GroupService) JoinGroup(ctx context.Context, userID, groupID uint64) error {
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}

	return s.groups.AddMember(ctx, groupID, userID)
}

func (s *GroupService) ListMyGroups(ctx context.Context, userID uint64) ([]model.GroupConversationSummary, error) {
	return s.groups.ListByUserIDWithLastMessage(ctx, userID)
}

func (s *GroupService) ListMembers(ctx context.Context, currentUserID, groupID uint64) ([]repository.GroupMemberProfile, error) {
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	isMember, err := s.groups.IsMember(ctx, groupID, currentUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	return s.groups.ListMembersWithProfile(ctx, groupID)
}

func (s *GroupService) LeaveGroup(ctx context.Context, currentUserID, groupID uint64) error {
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}

	isMember, err := s.groups.IsMember(ctx, groupID, currentUserID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotGroupMember
	}

	if group.OwnerID == currentUserID {
		return ErrGroupOwnerCannotLeave
	}

	removed, err := s.groups.RemoveMember(ctx, groupID, currentUserID)
	if err != nil {
		return err
	}
	if !removed {
		return ErrNotGroupMember
	}

	return nil
}

func (s *GroupService) RemoveMember(ctx context.Context, currentUserID, groupID, targetUserID uint64) error {
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}

	if group.OwnerID != currentUserID {
		return ErrNotGroupOwner
	}

	if targetUserID == group.OwnerID {
		return ErrCannotRemoveGroupOwner
	}

	isTargetMember, err := s.groups.IsMember(ctx, groupID, targetUserID)
	if err != nil {
		return err
	}
	if !isTargetMember {
		return ErrTargetNotGroupMember
	}

	removed, err := s.groups.RemoveMember(ctx, groupID, targetUserID)
	if err != nil {
		return err
	}
	if !removed {
		return ErrTargetNotGroupMember
	}

	return nil
}
