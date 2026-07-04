package service

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

var ErrNotificationNotFound = errors.New("notification not found")

type NotificationService struct{ repo repository.NotificationRepository }

func NewNotificationService(repo repository.NotificationRepository) *NotificationService { return &NotificationService{repo: repo} }

func (s *NotificationService) List(recipientID uuid.UUID, page, limit int) ([]dto.NotificationResponse, int64, error) {
	if recipientID == uuid.Nil { return nil, 0, ErrInvalidInput }
	page, limit = normalizePagination(page, limit)
	notifs, total, err := s.repo.ListByRecipient(recipientID, page, limit)
	if err != nil { return nil, 0, err }
	out := make([]dto.NotificationResponse, 0, len(notifs))
	for _, n := range notifs {
		out = append(out, buildNotificationResponse(n))
	}
	return out, total, nil
}

func (s *NotificationService) MarkRead(recipientID, notificationID uuid.UUID) error {
	if recipientID == uuid.Nil || notificationID == uuid.Nil { return ErrInvalidInput }
	if err := s.repo.MarkRead(notificationID, recipientID); errors.Is(err, gorm.ErrRecordNotFound) { return ErrNotificationNotFound } else if err != nil { return err }
	return nil
}

func buildNotificationResponse(n model.Notification) dto.NotificationResponse {
	var ref *string
	if n.ReferenceID != nil { v := n.ReferenceID.String(); ref = &v }
	return dto.NotificationResponse{ID: n.ID.String(), RecipientID: n.RecipientID.String(), ActorID: n.ActorID.String(), Type: n.Type, ReferenceID: ref, IsRead: n.IsRead, CreatedAt: n.CreatedAt}
}
