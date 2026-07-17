package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type DMRepository interface {
	FindConversationBetween(ctx context.Context, userA, userB uuid.UUID) (*model.Conversation, error)
	CreateConversation(ctx context.Context, conversation *model.Conversation) error
	AddParticipant(ctx context.Context, participant *model.ConversationParticipant) error
	ListConversations(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error)
	ListMessageRequests(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error)
	GetConversationParticipants(ctx context.Context, conversationIDs []uuid.UUID) ([]model.ConversationParticipant, error)
	ListMessages(ctx context.Context, conversationID uuid.UUID, page, limit int) ([]model.Message, int64, error)
	CreateMessage(ctx context.Context, message *model.Message) error
	UserBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	ConversationHasParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	FindConversationByID(ctx context.Context, id uuid.UUID) (*model.Conversation, error)
	GetStoryOwner(ctx context.Context, storyID uuid.UUID) (uuid.UUID, error)
	CreateNotification(ctx context.Context, notification *model.Notification) error
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*model.Message, error)
}

type GormDMRepository struct{ db *gorm.DB }

func NewDMRepository(db *gorm.DB) DMRepository { return &GormDMRepository{db: db} }

func (r *GormDMRepository) FindConversationBetween(ctx context.Context, userA, userB uuid.UUID) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.WithContext(ctx).
		Table("conversations").
		Select("conversations.*").
		Joins("JOIN conversation_participants cp1 ON cp1.conversation_id = conversations.id").
		Joins("JOIN conversation_participants cp2 ON cp2.conversation_id = conversations.id").
		Where("cp1.user_id = ? AND cp2.user_id = ? AND conversations.is_group = false", userA, userB).
		First(&conv).Error
	return &conv, err
}

func (r *GormDMRepository) CreateConversation(ctx context.Context, conversation *model.Conversation) error {
	return r.db.WithContext(ctx).Create(conversation).Error
}
func (r *GormDMRepository) AddParticipant(ctx context.Context, participant *model.ConversationParticipant) error {
	return r.db.WithContext(ctx).Create(participant).Error
}

func (r *GormDMRepository) ListMessageRequests(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error) {
	var total int64
	subQuery := r.db.Table("conversation_participants cp_other").
		Select("1").
		Joins("JOIN users u_other ON u_other.id = cp_other.user_id").
		Where("cp_other.conversation_id = conversations.id AND cp_other.user_id != ?", userID).
		Where("u_other.is_private = true").
		Where("NOT EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = ? AND f.following_id = cp_other.user_id AND f.status = 'accepted')", userID)

	blockedSubQuery := r.db.Table("conversation_participants cp_blocked").
		Select("1").
		Where("cp_blocked.conversation_id = conversations.id AND cp_blocked.user_id != ?", userID).
		Where("EXISTS (SELECT 1 FROM blocks b WHERE (b.blocker_id = ? AND b.blocked_id = cp_blocked.user_id) OR (b.blocker_id = cp_blocked.user_id AND b.blocked_id = ?))", userID, userID)

	query := r.db.WithContext(ctx).Model(&model.Conversation{}).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", userID).
		Where("EXISTS (?)", subQuery).
		Where("NOT EXISTS (?)", blockedSubQuery)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var convs []model.Conversation
	err := query.Order("conversations.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&convs).Error
	return convs, total, err
}

func (r *GormDMRepository) ListConversations(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error) {
	privateRequestSubQuery := r.db.Table("conversation_participants cp_other2").
		Select("1").
		Joins("JOIN users u_other2 ON u_other2.id = cp_other2.user_id").
		Where("cp_other2.conversation_id = conversations.id AND cp_other2.user_id != ?", userID).
		Where("u_other2.is_private = true").
		Where("NOT EXISTS (SELECT 1 FROM follows f2 WHERE f2.follower_id = ? AND f2.following_id = cp_other2.user_id AND f2.status = 'accepted')", userID)

	blockedSubQuery := r.db.Table("conversation_participants cp_blocked").
		Select("1").
		Where("cp_blocked.conversation_id = conversations.id AND cp_blocked.user_id != ?", userID).
		Where("EXISTS (SELECT 1 FROM blocks b WHERE (b.blocker_id = ? AND b.blocked_id = cp_blocked.user_id) OR (b.blocker_id = cp_blocked.user_id AND b.blocked_id = ?))", userID, userID)

	var total int64
	query := r.db.WithContext(ctx).Model(&model.Conversation{}).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", userID).
		Where("NOT EXISTS (?)", privateRequestSubQuery).
		Where("NOT EXISTS (?)", blockedSubQuery)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var convs []model.Conversation
	err := query.Order("conversations.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&convs).Error
	return convs, total, err
}

func (r *GormDMRepository) GetConversationParticipants(ctx context.Context, conversationIDs []uuid.UUID) ([]model.ConversationParticipant, error) {
	var parts []model.ConversationParticipant
	if len(conversationIDs) == 0 {
		return parts, nil
	}
	err := r.db.WithContext(ctx).Where("conversation_id IN ?", conversationIDs).Preload("User").Find(&parts).Error
	return parts, err
}

func (r *GormDMRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, page, limit int) ([]model.Message, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.Message{}).Where("conversation_id = ?", conversationID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var msgs []model.Message
	err := query.Preload("Sender").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&msgs).Error
	return msgs, total, err
}

func (r *GormDMRepository) CreateMessage(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *GormDMRepository) UserBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Block{}).Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userA, userB, userB, userA).Count(&count).Error
	return count > 0, err
}

func (r *GormDMRepository) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").Count(&count).Error
	return count > 0, err
}

func (r *GormDMRepository) ConversationHasParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ConversationParticipant{}).Where("conversation_id = ? AND user_id = ?", conversationID, userID).Count(&count).Error
	return count > 0, err
}

func (r *GormDMRepository) FindConversationByID(ctx context.Context, id uuid.UUID) (*model.Conversation, error) {
	var c model.Conversation
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormDMRepository) GetStoryOwner(ctx context.Context, storyID uuid.UUID) (uuid.UUID, error) {
	var s model.Story
	if err := r.db.WithContext(ctx).Select("user_id").First(&s, "id = ?", storyID).Error; err != nil {
		return uuid.Nil, err
	}
	return s.UserID, nil
}

func (r *GormDMRepository) CreateNotification(ctx context.Context, notification *model.Notification) error {
	return CreateNotificationHelper(ctx, r.db, notification)
}

func (r *GormDMRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*model.Message, error) {
	var m model.Message
	if err := r.db.WithContext(ctx).First(&m, "id = ?", messageID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *GormDMRepository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", messageID).Delete(&model.Message{}).Error
}

