package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
)

type fakeNotifRepo struct {
	notifs       map[uuid.UUID]*model.Notification
	readAll      bool
	notifCreated bool
}

func (f *fakeNotifRepo) ListByRecipient(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]model.Notification, int64, error) {
	var list []model.Notification
	for _, n := range f.notifs {
		if n.RecipientID == recipientID {
			list = append(list, *n)
		}
	}
	return list, int64(len(list)), nil
}

func (f *fakeNotifRepo) MarkRead(ctx context.Context, notificationID, recipientID uuid.UUID) error {
	if n, ok := f.notifs[notificationID]; ok && n.RecipientID == recipientID {
		n.IsRead = true
		return nil
	}
	return errors.New("not found")
}

func (f *fakeNotifRepo) MarkAllRead(ctx context.Context, recipientID uuid.UUID) error {
	f.readAll = true
	for _, n := range f.notifs {
		if n.RecipientID == recipientID {
			n.IsRead = true
		}
	}
	return nil
}

func (f *fakeNotifRepo) CreateNotification(ctx context.Context, notification *model.Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	if f.notifs == nil {
		f.notifs = make(map[uuid.UUID]*model.Notification)
	}
	f.notifs[notification.ID] = notification
	f.notifCreated = true
	return nil
}

func TestMarkAllNotificationsRead(t *testing.T) {
	recipientID := uuid.New()
	notifID := uuid.New()
	repo := &fakeNotifRepo{
		notifs: map[uuid.UUID]*model.Notification{
			notifID: {ID: notifID, RecipientID: recipientID, IsRead: false},
		},
	}
	svc := NewNotificationService(repo)

	err := svc.MarkAllRead(context.Background(), recipientID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !repo.readAll || !repo.notifs[notifID].IsRead {
		t.Fatal("expected all notifications marked as read")
	}
}

func TestCreateNotification(t *testing.T) {
	actorID := uuid.New()
	recipientID := uuid.New()
	repo := &fakeNotifRepo{}
	svc := NewNotificationService(repo)

	req := dto.CreateNotificationRequest{
		RecipientID: recipientID.String(),
		ActorID:     actorID.String(),
		Type:        "like",
	}

	err := svc.Create(context.Background(), actorID, req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !repo.notifCreated {
		t.Fatal("expected notification created")
	}
}
