package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type DMRepository interface {
	FindConversationBetween(userA, userB uuid.UUID) (*model.Conversation, error)
	CreateConversation(conversation *model.Conversation) error
	AddParticipant(participant *model.ConversationParticipant) error
	ListConversations(userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error)
	ListMessageRequests(userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error)
	GetConversationParticipants(conversationIDs []uuid.UUID) ([]model.ConversationParticipant, error)
	ListMessages(conversationID uuid.UUID, page, limit int) ([]model.Message, int64, error)
	CreateMessage(message *model.Message) error
	UserBlocked(userA, userB uuid.UUID) (bool, error)
	IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error)
	ConversationHasParticipant(conversationID, userID uuid.UUID) (bool, error)
	FindConversationByID(id uuid.UUID) (*model.Conversation, error)
	GetStoryOwner(storyID uuid.UUID) (uuid.UUID, error)
	CreateNotification(notification *model.Notification) error
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

func (r *GormDMRepository) ListMessageRequests(userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error) {
	// MSG-02: conversations where the OTHER user is private
	// and current user is NOT an accepted follower of the other user
	var total int64
	subQuery := r.db.Table("conversation_participants cp_other").
		Select("1").
		Joins("JOIN users u_other ON u_other.id = cp_other.user_id").
		Where("cp_other.conversation_id = conversations.id AND cp_other.user_id != ?", userID).
		Where("u_other.is_private = true").
		Where("NOT EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = ? AND f.following_id = cp_other.user_id AND f.status = 'accepted')", userID)

	query := r.db.Model(&model.Conversation{}).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", userID).
		Where("EXISTS (?)", subQuery)
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var convs []model.Conversation
	err := query.Order("conversations.created_at DESC").Offset((page-1)*limit).Limit(limit).Find(&convs).Error
	return convs, total, err
}

func (r *GormDMRepository) ListConversations(userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error) {
	// MSG-02: exclude conversations with private-account users where current user is NOT an accepted follower
	privateRequestSubQuery := r.db.Table("conversation_participants cp_other2").
		Select("1").
		Joins("JOIN users u_other2 ON u_other2.id = cp_other2.user_id").
		Where("cp_other2.conversation_id = conversations.id AND cp_other2.user_id != ?", userID).
		Where("u_other2.is_private = true").
		Where("NOT EXISTS (SELECT 1 FROM follows f2 WHERE f2.follower_id = ? AND f2.following_id = cp_other2.user_id AND f2.status = 'accepted')", userID)

	var total int64
	query := r.db.Model(&model.Conversation{}).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", userID).
		Where("NOT EXISTS (?)", privateRequestSubQuery)
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

func (r *GormDMRepository) GetStoryOwner(storyID uuid.UUID) (uuid.UUID, error) {
	var s model.Story
	if err := r.db.Select("user_id").First(&s, "id = ?", storyID).Error; err != nil { return uuid.Nil, err }
	return s.UserID, nil
}

func (r *GormDMRepository) CreateNotification(notification *model.Notification) error {
	return r.db.Create(notification).Error
}
