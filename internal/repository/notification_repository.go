package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type NotificationRepository interface {
	ListByRecipient(recipientID uuid.UUID, page, limit int) ([]model.Notification, int64, error)
	MarkRead(notificationID, recipientID uuid.UUID) error
	MarkAllRead(recipientID uuid.UUID) error
	CreateNotification(notification *model.Notification) error
}

type GormNotificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (r *GormNotificationRepository) ListByRecipient(recipientID uuid.UUID, page, limit int) ([]model.Notification, int64, error) {
	var total int64
	// NTF-02: exclude notifications from blocked users (both directions)
	blockedFilter := `actor_id NOT IN (
		SELECT blocker_id FROM blocks WHERE blocked_id = ?
		UNION
		SELECT blocked_id FROM blocks WHERE blocker_id = ?
	)`
	query := r.db.Model(&model.Notification{}).Where("recipient_id = ? AND "+blockedFilter, recipientID, recipientID, recipientID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var notifs []model.Notification
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Preload("Actor").Find(&notifs).Error
	return notifs, total, err
}

func (r *GormNotificationRepository) MarkRead(notificationID, recipientID uuid.UUID) error {
	res := r.db.Model(&model.Notification{}).Where("id = ? AND recipient_id = ?", notificationID, recipientID).Update("is_read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormNotificationRepository) MarkAllRead(recipientID uuid.UUID) error {
	return r.db.Model(&model.Notification{}).Where("recipient_id = ? AND is_read = false", recipientID).Update("is_read", true).Error
}

func (r *GormNotificationRepository) CreateNotification(notification *model.Notification) error {
	return r.db.Create(notification).Error
}
