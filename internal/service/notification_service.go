package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

var ErrNotificationNotFound = errors.New("notification not found")

type NotificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) List(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]dto.NotificationResponse, int64, error) {
	if recipientID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	page, limit = normalizePagination(page, limit)
	notifs, total, err := s.repo.ListByRecipient(ctx, recipientID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.NotificationResponse, 0, len(notifs))
	for _, n := range notifs {
		out = append(out, buildNotificationResponse(n))
	}
	return out, total, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, recipientID, notificationID uuid.UUID) error {
	if recipientID == uuid.Nil || notificationID == uuid.Nil {
		return ErrInvalidInput
	}
	if err := s.repo.MarkRead(ctx, notificationID, recipientID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotificationNotFound
	} else if err != nil {
		return err
	}
	return nil
}

func (s *NotificationService) MarkAllRead(ctx context.Context, recipientID uuid.UUID) error {
	if recipientID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.MarkAllRead(ctx, recipientID)
}

func (s *NotificationService) Create(ctx context.Context, actorID uuid.UUID, req dto.CreateNotificationRequest) error {
	if actorID == uuid.Nil {
		return ErrInvalidInput
	}
	recID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		return errors.New("invalid recipient_id")
	}
	var refID *uuid.UUID
	if req.ReferenceID != nil {
		tmp, err := uuid.Parse(*req.ReferenceID)
		if err == nil {
			refID = &tmp
		}
	}
	notif := &model.Notification{RecipientID: recID, ActorID: actorID, Type: req.Type, ReferenceID: refID}
	return s.repo.CreateNotification(ctx, notif)
}

func buildNotificationResponse(n model.Notification) dto.NotificationResponse {
	var ref *string
	if n.ReferenceID != nil {
		v := n.ReferenceID.String()
		ref = &v
	}
	return dto.NotificationResponse{ID: n.ID.String(), RecipientID: n.RecipientID.String(), ActorID: n.ActorID.String(), Type: n.Type, ReferenceID: ref, IsRead: n.IsRead, CreatedAt: n.CreatedAt}
}
