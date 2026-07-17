package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/model"
)

type NotificationRepository interface {
	ListByRecipient(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]model.Notification, int64, error)
	MarkRead(ctx context.Context, notificationID, recipientID uuid.UUID) error
	MarkAllRead(ctx context.Context, recipientID uuid.UUID) error
	CreateNotification(ctx context.Context, notification *model.Notification) error
}

type GormNotificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &GormNotificationRepository{db: db}
}

func CreateNotificationHelper(ctx context.Context, db *gorm.DB, n *model.Notification) error {
	if n == nil || n.ActorID == n.RecipientID {
		return nil // NTF-01: self-notification prevention
	}
	var count int64
	err := db.WithContext(ctx).Model(&model.Block{}).Where(
		"(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
		n.ActorID, n.RecipientID, n.RecipientID, n.ActorID,
	).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // NTF-02: suppress notifications from blocked users
	}
	return db.WithContext(ctx).Create(n).Error
}

func (r *GormNotificationRepository) ListByRecipient(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]model.Notification, int64, error) {
	var total int64
	// NTF-02: exclude notifications from blocked users (both directions)
	blockedFilter := `actor_id NOT IN (
		SELECT blocker_id FROM blocks WHERE blocked_id = ?
		UNION
		SELECT blocked_id FROM blocks WHERE blocker_id = ?
	)`
	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("recipient_id = ? AND "+blockedFilter, recipientID, recipientID, recipientID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var notifs []model.Notification
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Preload("Actor").Find(&notifs).Error
	return notifs, total, err
}

func (r *GormNotificationRepository) MarkRead(ctx context.Context, notificationID, recipientID uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ? AND recipient_id = ?", notificationID, recipientID).Update("is_read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormNotificationRepository) MarkAllRead(ctx context.Context, recipientID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("recipient_id = ? AND is_read = false", recipientID).Update("is_read", true).Error
}

func (r *GormNotificationRepository) CreateNotification(ctx context.Context, notification *model.Notification) error {
	return CreateNotificationHelper(ctx, r.db, notification)
}
