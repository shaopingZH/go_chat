package repository

import (
	"context"
	"errors"
	"time"

	"go-chat/internal/model"
	"gorm.io/gorm"
)

var ErrInvalidHistoryCursor = errors.New("invalid last_msg_id")
var ErrInvalidHistoryAnchor = errors.New("invalid around_msg_id")

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	ListPrivateHistory(ctx context.Context, userID, targetID, lastMsgID uint64, limit int, keyword string) ([]model.Message, bool, error)
	ListGroupHistory(ctx context.Context, groupID, lastMsgID uint64, limit int, keyword string) ([]model.Message, bool, error)
	ListPrivateHistoryAround(ctx context.Context, userID, targetID, aroundMsgID uint64, limit int) ([]model.Message, bool, error)
	ListGroupHistoryAround(ctx context.Context, groupID, aroundMsgID uint64, limit int) ([]model.Message, bool, error)
	ListPrivateConversations(ctx context.Context, userID uint64, limit int) ([]model.PrivateConversationSummary, error)
	GetReadState(ctx context.Context, userID uint64, chatType string, targetID uint64) (*model.ConversationReadState, error)
	UpsertReadState(ctx context.Context, userID uint64, chatType string, targetID uint64, lastReadMessageID uint64) error
	GetLatestPrivateMessageID(ctx context.Context, userID, targetID uint64) (uint64, error)
	GetLatestGroupMessageID(ctx context.Context, groupID uint64) (uint64, error)
	CountUnreadPrivate(ctx context.Context, userID, targetID, lastReadMessageID uint64) (int64, error)
	CountUnreadGroup(ctx context.Context, userID, groupID, lastReadMessageID uint64) (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) ListPrivateHistory(ctx context.Context, userID, targetID, lastMsgID uint64, limit int, keyword string) ([]model.Message, bool, error) {
	query := r.db.WithContext(ctx).
		Where("group_id IS NULL").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, targetID, targetID, userID)

	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	if lastMsgID > 0 {
		var exists int64
		err := r.db.WithContext(ctx).
			Model(&model.Message{}).
			Where("id = ?", lastMsgID).
			Where("group_id IS NULL").
			Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, targetID, targetID, userID).
			Count(&exists).Error
		if err != nil {
			return nil, false, err
		}
		if exists == 0 {
			return nil, false, ErrInvalidHistoryCursor
		}
		query = query.Where("id < ?", lastMsgID)
	}

	var messages []model.Message
	err := query.Order("id DESC").Limit(limit + 1).Find(&messages).Error
	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	reverseMessages(messages)
	return messages, hasMore, nil
}

func (r *messageRepository) ListGroupHistory(ctx context.Context, groupID, lastMsgID uint64, limit int, keyword string) ([]model.Message, bool, error) {
	query := r.db.WithContext(ctx).
		Where("group_id = ?", groupID)

	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	if lastMsgID > 0 {
		var exists int64
		err := r.db.WithContext(ctx).
			Model(&model.Message{}).
			Where("id = ?", lastMsgID).
			Where("group_id = ?", groupID).
			Count(&exists).Error
		if err != nil {
			return nil, false, err
		}
		if exists == 0 {
			return nil, false, ErrInvalidHistoryCursor
		}
		query = query.Where("id < ?", lastMsgID)
	}

	var messages []model.Message
	err := query.Order("id DESC").Limit(limit + 1).Find(&messages).Error
	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	reverseMessages(messages)
	return messages, hasMore, nil
}

func (r *messageRepository) ListPrivateHistoryAround(ctx context.Context, userID, targetID, aroundMsgID uint64, limit int) ([]model.Message, bool, error) {
	anchor, err := fetchAnchorMessage(r.privateHistoryBaseQuery(ctx, userID, targetID), aroundMsgID)
	if err != nil {
		return nil, false, err
	}

	olderCandidates, newerCandidates, err := fetchContextCandidates(
		r.privateHistoryBaseQuery(ctx, userID, targetID),
		r.privateHistoryBaseQuery(ctx, userID, targetID),
		aroundMsgID,
		limit,
	)
	if err != nil {
		return nil, false, err
	}

	return buildAroundHistoryPage(anchor, olderCandidates, newerCandidates, limit), len(olderCandidates) > selectAroundCounts(limit, len(olderCandidates), len(newerCandidates)).before, nil
}

func (r *messageRepository) ListGroupHistoryAround(ctx context.Context, groupID, aroundMsgID uint64, limit int) ([]model.Message, bool, error) {
	anchor, err := fetchAnchorMessage(r.groupHistoryBaseQuery(ctx, groupID), aroundMsgID)
	if err != nil {
		return nil, false, err
	}

	olderCandidates, newerCandidates, err := fetchContextCandidates(
		r.groupHistoryBaseQuery(ctx, groupID),
		r.groupHistoryBaseQuery(ctx, groupID),
		aroundMsgID,
		limit,
	)
	if err != nil {
		return nil, false, err
	}

	return buildAroundHistoryPage(anchor, olderCandidates, newerCandidates, limit), len(olderCandidates) > selectAroundCounts(limit, len(olderCandidates), len(newerCandidates)).before, nil
}

func (r *messageRepository) ListPrivateConversations(ctx context.Context, userID uint64, limit int) ([]model.PrivateConversationSummary, error) {
	query := `
SELECT
  f.friend_id AS partner_id,
  COALESCE(u.username, '') AS partner_username,
  COALESCE(u.avatar, '') AS partner_avatar,
  COALESCE((
    SELECT COUNT(1)
    FROM messages AS um
    WHERE um.group_id IS NULL
      AND um.sender_id = f.friend_id
      AND um.receiver_id = f.user_id
      AND um.id > COALESCE(cr.last_read_message_id, 0)
  ), 0) AS unread_count,
  m.id AS last_message_id,
  m.sender_id AS last_sender_id,
  m.msg_type AS last_msg_type,
  COALESCE(m.content, '') AS last_content,
  m.created_at AS last_created_at
FROM friendships AS f
LEFT JOIN users AS u ON u.id = f.friend_id
LEFT JOIN conversation_read_states AS cr
  ON cr.user_id = f.user_id
 AND cr.chat_type = 'private'
 AND cr.target_id = f.friend_id
LEFT JOIN (
  SELECT
    t.partner_id,
    MAX(t.id) AS max_id
  FROM (
    SELECT
      CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS partner_id,
      id
    FROM messages
    WHERE group_id IS NULL
      AND (sender_id = ? OR receiver_id = ?)
  ) AS t
  GROUP BY t.partner_id
) AS lm ON lm.partner_id = f.friend_id
LEFT JOIN messages AS m ON m.id = lm.max_id
WHERE f.user_id = ?
ORDER BY (m.id IS NULL) ASC, m.id DESC, f.created_at DESC
LIMIT ?
`

	var conversations []model.PrivateConversationSummary
	err := r.db.WithContext(ctx).
		Raw(query, userID, userID, userID, userID, limit).
		Scan(&conversations).Error
	if err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *messageRepository) GetReadState(ctx context.Context, userID uint64, chatType string, targetID uint64) (*model.ConversationReadState, error) {
	var state model.ConversationReadState
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND chat_type = ? AND target_id = ?", userID, chatType, targetID).
		First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *messageRepository) UpsertReadState(ctx context.Context, userID uint64, chatType string, targetID uint64, lastReadMessageID uint64) error {
	now := time.Now().UTC()
	query := `
INSERT INTO conversation_read_states (user_id, chat_type, target_id, last_read_message_id, last_read_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  last_read_message_id = GREATEST(last_read_message_id, VALUES(last_read_message_id)),
  last_read_at = VALUES(last_read_at),
  updated_at = VALUES(updated_at)
`
	return r.db.WithContext(ctx).
		Exec(query, userID, chatType, targetID, lastReadMessageID, now, now, now).Error
}

func (r *messageRepository) GetLatestPrivateMessageID(ctx context.Context, userID, targetID uint64) (uint64, error) {
	var latestID uint64
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("group_id IS NULL").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, targetID, targetID, userID).
		Order("id DESC").
		Limit(1).
		Pluck("id", &latestID).Error
	if err != nil {
		return 0, err
	}
	return latestID, nil
}

func (r *messageRepository) GetLatestGroupMessageID(ctx context.Context, groupID uint64) (uint64, error) {
	var latestID uint64
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("group_id = ?", groupID).
		Order("id DESC").
		Limit(1).
		Pluck("id", &latestID).Error
	if err != nil {
		return 0, err
	}
	return latestID, nil
}

func (r *messageRepository) CountUnreadPrivate(ctx context.Context, userID, targetID, lastReadMessageID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("group_id IS NULL").
		Where("sender_id = ? AND receiver_id = ?", targetID, userID).
		Where("id > ?", lastReadMessageID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *messageRepository) CountUnreadGroup(ctx context.Context, userID, groupID, lastReadMessageID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("group_id = ?", groupID).
		Where("sender_id <> ?", userID).
		Where("id > ?", lastReadMessageID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func reverseMessages(messages []model.Message) {
	left, right := 0, len(messages)-1
	for left < right {
		messages[left], messages[right] = messages[right], messages[left]
		left++
		right--
	}
}

func (r *messageRepository) privateHistoryBaseQuery(ctx context.Context, userID, targetID uint64) *gorm.DB {
	return r.db.WithContext(ctx).
		Where("group_id IS NULL").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, targetID, targetID, userID)
}

func (r *messageRepository) groupHistoryBaseQuery(ctx context.Context, groupID uint64) *gorm.DB {
	return r.db.WithContext(ctx).
		Where("group_id = ?", groupID)
}

func fetchAnchorMessage(baseQuery *gorm.DB, aroundMsgID uint64) (model.Message, error) {
	var anchor model.Message
	err := baseQuery.Where("id = ?", aroundMsgID).First(&anchor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Message{}, ErrInvalidHistoryAnchor
	}
	if err != nil {
		return model.Message{}, err
	}
	return anchor, nil
}

func fetchContextCandidates(olderQuery, newerQuery *gorm.DB, aroundMsgID uint64, limit int) ([]model.Message, []model.Message, error) {
	windowSize := limit - 1
	if windowSize < 1 {
		windowSize = 1
	}
	olderCandidates := make([]model.Message, 0, windowSize)
	if err := olderQuery.Where("id < ?", aroundMsgID).Order("id DESC").Limit(windowSize).Find(&olderCandidates).Error; err != nil {
		return nil, nil, err
	}

	if limit <= 1 {
		return olderCandidates, nil, nil
	}

	newerCandidates := make([]model.Message, 0, windowSize)
	if err := newerQuery.Where("id > ?", aroundMsgID).Order("id ASC").Limit(windowSize).Find(&newerCandidates).Error; err != nil {
		return nil, nil, err
	}

	return olderCandidates, newerCandidates, nil
}

type aroundCounts struct {
	before int
	after  int
}

func selectAroundCounts(limit, olderAvailable, newerAvailable int) aroundCounts {
	if limit <= 1 {
		return aroundCounts{}
	}

	beforeTarget := (limit - 1) / 2
	afterTarget := (limit - 1) - beforeTarget

	before := minInt(olderAvailable, beforeTarget)
	after := minInt(newerAvailable, afterTarget)
	remaining := (limit - 1) - before - after

	if remaining > 0 && olderAvailable > before {
		extra := minInt(remaining, olderAvailable-before)
		before += extra
		remaining -= extra
	}
	if remaining > 0 && newerAvailable > after {
		extra := minInt(remaining, newerAvailable-after)
		after += extra
	}

	return aroundCounts{before: before, after: after}
}

func buildAroundHistoryPage(anchor model.Message, olderCandidates, newerCandidates []model.Message, limit int) []model.Message {
	counts := selectAroundCounts(limit, len(olderCandidates), len(newerCandidates))
	older := append([]model.Message(nil), olderCandidates[:counts.before]...)
	reverseMessages(older)

	items := make([]model.Message, 0, counts.before+1+counts.after)
	items = append(items, older...)
	items = append(items, anchor)
	items = append(items, newerCandidates[:counts.after]...)
	return items
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
