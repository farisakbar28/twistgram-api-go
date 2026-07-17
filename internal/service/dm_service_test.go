package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
)

type fakeDMRepo struct {
	blocked       bool
	convs         map[uuid.UUID]*model.Conversation
	participants  map[uuid.UUID][]uuid.UUID
	msgs          map[uuid.UUID][]model.Message
	storyOwner    uuid.UUID
	notifSent     bool
}

func (f *fakeDMRepo) FindConversationBetween(ctx context.Context, userA, userB uuid.UUID) (*model.Conversation, error) {
	for cid, parts := range f.participants {
		hasA, hasB := false, false
		for _, uid := range parts {
			if uid == userA {
				hasA = true
			}
			if uid == userB {
				hasB = true
			}
		}
		if hasA && hasB {
			return f.convs[cid], nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeDMRepo) CreateConversation(ctx context.Context, conversation *model.Conversation) error {
	if conversation.ID == uuid.Nil {
		conversation.ID = uuid.New()
	}
	if f.convs == nil {
		f.convs = make(map[uuid.UUID]*model.Conversation)
	}
	f.convs[conversation.ID] = conversation
	return nil
}

func (f *fakeDMRepo) AddParticipant(ctx context.Context, participant *model.ConversationParticipant) error {
	if f.participants == nil {
		f.participants = make(map[uuid.UUID][]uuid.UUID)
	}
	f.participants[participant.ConversationID] = append(f.participants[participant.ConversationID], participant.UserID)
	return nil
}

func (f *fakeDMRepo) ListConversations(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error) {
	return nil, 0, nil
}

func (f *fakeDMRepo) ListMessageRequests(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Conversation, int64, error) {
	return nil, 0, nil
}

func (f *fakeDMRepo) GetConversationParticipants(ctx context.Context, conversationIDs []uuid.UUID) ([]model.ConversationParticipant, error) {
	return nil, nil
}

func (f *fakeDMRepo) ListMessages(ctx context.Context, conversationID uuid.UUID, page, limit int) ([]model.Message, int64, error) {
	list := f.msgs[conversationID]
	return list, int64(len(list)), nil
}

func (f *fakeDMRepo) CreateMessage(ctx context.Context, message *model.Message) error {
	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}
	if f.msgs == nil {
		f.msgs = make(map[uuid.UUID][]model.Message)
	}
	f.msgs[message.ConversationID] = append(f.msgs[message.ConversationID], *message)
	return nil
}

func (f *fakeDMRepo) UserBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	return f.blocked, nil
}

func (f *fakeDMRepo) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return false, nil
}

func (f *fakeDMRepo) ConversationHasParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	parts, ok := f.participants[conversationID]
	if !ok {
		return false, nil
	}
	for _, p := range parts {
		if p == userID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeDMRepo) FindConversationByID(ctx context.Context, id uuid.UUID) (*model.Conversation, error) {
	c, ok := f.convs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (f *fakeDMRepo) GetStoryOwner(ctx context.Context, storyID uuid.UUID) (uuid.UUID, error) {
	return f.storyOwner, nil
}

func (f *fakeDMRepo) CreateNotification(ctx context.Context, notification *model.Notification) error {
	f.notifSent = true
	return nil
}

func TestStartConversationBlockedFails(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	repo := &fakeDMRepo{blocked: true}
	svc := NewDMService(repo)

	_, err := svc.StartConversation(context.Background(), userA, userB)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for blocked DM start, got %v", err)
	}
}

func TestSendMessageWithStoryReply(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	convID := uuid.New()
	storyID := uuid.New()
	repo := &fakeDMRepo{
		participants: map[uuid.UUID][]uuid.UUID{convID: {userA, userB}},
		storyOwner:   userB,
	}
	svc := NewDMService(repo)

	content := "Cool story!"
	storyStr := storyID.String()
	res, err := svc.SendMessage(context.Background(), userA, convID, dto.MessageRequest{
		Content: &content,
		StoryID: &storyStr,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.ReplyToStoryID == nil || *res.ReplyToStoryID != storyStr {
		t.Fatalf("expected ReplyToStoryID %s, got %v", storyStr, res.ReplyToStoryID)
	}
	if !repo.notifSent {
		t.Fatal("expected notification sent to story owner")
	}
}
