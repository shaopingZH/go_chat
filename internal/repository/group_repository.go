package repository

import (
	"context"
	"errors"
	"time"

	"go-chat/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GroupRepository interface {
	Create(ctx context.Context, group *model.Group) error
	GetByID(ctx context.Context, groupID uint64) (*model.Group, error)
	UpdateProfileByID(ctx context.Context, groupID uint64, name, avatar string) error
	AddMember(ctx context.Context, groupID, userID uint64) error
	IsMember(ctx context.Context, groupID, userID uint64) (bool, error)
	ListByUserID(ctx context.Context, userID uint64) ([]model.Group, error)
	ListByUserIDWithLastMessage(ctx context.Context, userID uint64) ([]model.GroupConversationSummary, error)
	ListMemberIDs(ctx context.Context, groupID uint64) ([]uint64, error)
	ListMembersWithProfile(ctx context.Context, groupID uint64) ([]GroupMemberProfile, error)
	RemoveMember(ctx context.Context, groupID, userID uint64) (bool, error)
}

type GroupMemberProfile struct {
	ID          uint64    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Avatar      string    `json:"avatar"`
	JoinedAt    time.Time `json:"joined_at"`
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *groupRepository) GetByID(ctx context.Context, groupID uint64) (*model.Group, error) {
	var group model.Group
	err := r.db.WithContext(ctx).First(&group, groupID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *groupRepository) UpdateProfileByID(ctx context.Context, groupID uint64, name, avatar string) error {
	return r.db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", groupID).
		Updates(map[string]any{
			"name":   name,
			"avatar": avatar,
		}).Error
}

func (r *groupRepository) AddMember(ctx context.Context, groupID, userID uint64) error {
	member := model.GroupMember{GroupID: groupID, UserID: userID, JoinedAt: time.Now()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(&member).Error
}

func (r *groupRepository) IsMember(ctx context.Context, groupID, userID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *groupRepository) ListByUserID(ctx context.Context, userID uint64) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.WithContext(ctx).
		Table("`groups` AS g").
		Select("g.*").
		Joins("JOIN group_members gm ON gm.group_id = g.id").
		Where("gm.user_id = ?", userID).
		Order("g.id DESC").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}

type groupWithLastMessageRow struct {
	ID                   uint64
	Name                 string
	Avatar               string
	OwnerID              uint64
	CreatedAt            time.Time
	UnreadCount          int64
	LastMessageID        *uint64
	LastMessageSenderID  *uint64
	LastMessageMsgType   *int8
	LastMessageContent   *string
	LastMessageCreatedAt *time.Time
}

func (r *groupRepository) ListByUserIDWithLastMessage(ctx context.Context, userID uint64) ([]model.GroupConversationSummary, error) {
	query := `
SELECT
  g.id,
  g.name,
  g.avatar,
  g.owner_id,
  g.created_at,
  COALESCE((
    SELECT COUNT(1)
    FROM messages AS um
    WHERE um.group_id = g.id
      AND um.sender_id <> gm.user_id
      AND um.id > COALESCE(cr.last_read_message_id, 0)
  ), 0) AS unread_count,
  lm.id AS last_message_id,
  lm.sender_id AS last_message_sender_id,
  lm.msg_type AS last_message_msg_type,
  lm.content AS last_message_content,
  lm.created_at AS last_message_created_at
FROM ` + "`groups`" + ` g
JOIN group_members gm ON gm.group_id = g.id
LEFT JOIN conversation_read_states cr
  ON cr.user_id = gm.user_id
 AND cr.chat_type = 'group'
 AND cr.target_id = g.id
LEFT JOIN (
  SELECT m1.group_id, m1.id, m1.sender_id, m1.msg_type, m1.content, m1.created_at
  FROM messages m1
  JOIN (
    SELECT group_id, MAX(id) AS max_id
    FROM messages
    WHERE group_id IS NOT NULL
    GROUP BY group_id
  ) latest ON latest.max_id = m1.id
) lm ON lm.group_id = g.id
WHERE gm.user_id = ?
ORDER BY g.id DESC
`

	var rows []groupWithLastMessageRow
	err := r.db.WithContext(ctx).Raw(query, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]model.GroupConversationSummary, 0, len(rows))
	for _, row := range rows {
		item := model.GroupConversationSummary{
			ID:          row.ID,
			Name:        row.Name,
			Avatar:      row.Avatar,
			OwnerID:     row.OwnerID,
			CreatedAt:   row.CreatedAt,
			UnreadCount: row.UnreadCount,
		}
		if row.LastMessageID != nil && row.LastMessageSenderID != nil && row.LastMessageMsgType != nil && row.LastMessageContent != nil && row.LastMessageCreatedAt != nil {
			item.LastMessage = &model.GroupLastMessageSummary{
				ID:        *row.LastMessageID,
				SenderID:  *row.LastMessageSenderID,
				MsgType:   *row.LastMessageMsgType,
				Content:   *row.LastMessageContent,
				CreatedAt: *row.LastMessageCreatedAt,
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *groupRepository) ListMemberIDs(ctx context.Context, groupID uint64) ([]uint64, error) {
	var memberIDs []uint64
	err := r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &memberIDs).Error
	if err != nil {
		return nil, err
	}

	return memberIDs, nil
}

func (r *groupRepository) ListMembersWithProfile(ctx context.Context, groupID uint64) ([]GroupMemberProfile, error) {
	var members []GroupMemberProfile
	err := r.db.WithContext(ctx).
		Table("group_members AS gm").
		Select("u.id, u.username, u.display_name, u.avatar, gm.joined_at").
		Joins("JOIN users u ON u.id = gm.user_id").
		Where("gm.group_id = ?", groupID).
		Order("gm.joined_at ASC").
		Scan(&members).Error
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID uint64) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&model.GroupMember{})
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
