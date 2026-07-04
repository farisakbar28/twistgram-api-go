package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"twistgram-api-go/internal/model"
)

type DMRepository interface {
	FindConversationBetween(userA, userB uuid.UUID) (*model.Conversation, error)
	CreateConversation(conversation *model.Conversation) error
	AddParticipant(participant *model.ConversationParticipant) error
	ListConversations(userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error)
	GetConversationParticipants(conversationIDs []uuid.UUID) ([]model.ConversationParticipant, error)
	ListMessages(conversationID uuid.UUID, page, limit int) ([]model.Message, int64, error)
	CreateMessage(message *model.Message) error
	UserBlocked(userA, userB uuid.UUID) (bool, error)
	IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error)
	ConversationHasParticipant(conversationID, userID uuid.UUID) (bool, error)
	FindConversationByID(id uuid.UUID) (*model.Conversation, error)
}

type GormDMRepository struct{ db *gorm.DB }

func NewDMRepository(db *gorm.DB) DMRepository { return &GormDMRepository{db: db} }

func (r *GormDMRepository) FindConversationBetween(userA, userB uuid.UUID) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.
		Table("conversations").
		Select("conversations.*").
		Joins("JOIN conversation_participants cp1 ON cp1.conversation_id = conversations.id").
		Joins("JOIN conversation_participants cp2 ON cp2.conversation_id = conversations.id").
		Where("cp1.user_id = ? AND cp2.user_id = ? AND conversations.is_group = false", userA, userB).
		First(&conv).Error
	return &conv, err
}

func (r *GormDMRepository) CreateConversation(conversation *model.Conversation) error { return r.db.Create(conversation).Error }
func (r *GormDMRepository) AddParticipant(participant *model.ConversationParticipant) error { return r.db.Create(participant).Error }

func (r *GormDMRepository) ListConversations(userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error) {
	var total int64
	query := r.db.Model(&model.Conversation{}).Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").Where("cp.user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var convs []model.Conversation
	err := query.Order("conversations.created_at DESC").Offset((page-1)*limit).Limit(limit).Find(&convs).Error
	return convs, total, err
}

func (r *GormDMRepository) GetConversationParticipants(conversationIDs []uuid.UUID) ([]model.ConversationParticipant, error) {
	var parts []model.ConversationParticipant
	if len(conversationIDs) == 0 { return parts, nil }
	err := r.db.Where("conversation_id IN ?", conversationIDs).Preload("User").Find(&parts).Error
	return parts, err
}

func (r *GormDMRepository) ListMessages(conversationID uuid.UUID, page, limit int) ([]model.Message, int64, error) {
	var total int64
	query := r.db.Model(&model.Message{}).Where("conversation_id = ?", conversationID)
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var msgs []model.Message
	err := query.Preload("Sender").Order("created_at DESC").Offset((page-1)*limit).Limit(limit).Find(&msgs).Error
	return msgs, total, err
}

func (r *GormDMRepository) CreateMessage(message *model.Message) error { return r.db.Create(message).Error }

func (r *GormDMRepository) UserBlocked(userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Block{}).Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userA, userB, userB, userA).Count(&count).Error
	return count > 0, err
}

func (r *GormDMRepository) IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).Where("follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").Count(&count).Error
	return count > 0, err
}

func (r *GormDMRepository) ConversationHasParticipant(conversationID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.ConversationParticipant{}).Where("conversation_id = ? AND user_id = ?", conversationID, userID).Count(&count).Error
	return count > 0, err
}

func (r *GormDMRepository) FindConversationByID(id uuid.UUID) (*model.Conversation, error) {
	var c model.Conversation
	if err := r.db.First(&c, "id = ?", id).Error; err != nil { return nil, err }
	return &c, nil
}

func (r *GormDMRepository) _unused(time.Time) {}
var _ = errors.New
var _ = clause.OnConflict{}
