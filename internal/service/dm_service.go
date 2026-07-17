package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

var ErrConversationNotFound = errors.New("conversation not found")

type DMService struct{ repo repository.DMRepository }

func NewDMService(repo repository.DMRepository) *DMService { return &DMService{repo: repo} }

func (s *DMService) StartConversation(ctx context.Context, userID, targetID uuid.UUID) (*dto.ConversationResponse, error) {
	if userID == uuid.Nil || targetID == uuid.Nil || userID == targetID {
		return nil, ErrInvalidInput
	}
	blocked, err := s.repo.UserBlocked(ctx, userID, targetID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrForbidden
	}
	if existing, err := s.repo.FindConversationBetween(ctx, userID, targetID); err == nil && existing != nil {
		return &dto.ConversationResponse{ID: existing.ID.String(), IsGroup: existing.IsGroup, CreatedAt: existing.CreatedAt}, nil
	}
	conv := &model.Conversation{}
	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}
	if err := s.repo.AddParticipant(ctx, &model.ConversationParticipant{ConversationID: conv.ID, UserID: userID}); err != nil {
		return nil, err
	}
	if err := s.repo.AddParticipant(ctx, &model.ConversationParticipant{ConversationID: conv.ID, UserID: targetID}); err != nil {
		return nil, err
	}
	return &dto.ConversationResponse{ID: conv.ID.String(), IsGroup: conv.IsGroup, CreatedAt: conv.CreatedAt}, nil
}

func (s *DMService) ListConversations(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.ConversationResponse, int64, error) {
	if userID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	pg, lm := normalizePagination(page, limit)
	convs, total, err := s.repo.ListConversations(ctx, userID, pg, lm)
	if err != nil {
		return nil, 0, err
	}

	convIDs := make([]uuid.UUID, 0, len(convs))
	for _, c := range convs {
		convIDs = append(convIDs, c.ID)
	}

	parts, err := s.repo.GetConversationParticipants(ctx, convIDs)
	if err != nil {
		return nil, 0, err
	}

	partsMap := make(map[uuid.UUID]*dto.SearchUserItem)
	for _, p := range parts {
		if p.UserID != userID && !p.Conversation.IsGroup {
			partsMap[p.ConversationID] = &dto.SearchUserItem{
				ID: p.User.ID.String(), Name: p.User.Name, Username: p.User.Username, AvatarURL: p.User.AvatarURL,
			}
		}
	}

	out := make([]dto.ConversationResponse, 0, len(convs))
	for _, c := range convs {
		res := dto.ConversationResponse{ID: c.ID.String(), IsGroup: c.IsGroup, CreatedAt: c.CreatedAt}
		if other, ok := partsMap[c.ID]; ok {
			res.OtherUser = other
		}
		out = append(out, res)
	}
	return out, total, nil
}

func (s *DMService) ListMessages(ctx context.Context, userID, conversationID uuid.UUID, page, limit int) ([]dto.MessageResponse, int64, error) {
	if userID == uuid.Nil || conversationID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	ok, err := s.repo.ConversationHasParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, ErrForbidden
	}
	pg, lm := normalizePagination(page, limit)
	msgs, total, err := s.repo.ListMessages(ctx, conversationID, pg, lm)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.MessageResponse, 0, len(msgs))
	for _, m := range msgs {
		var replyToStoryID *string
		if m.ReplyToStoryID != nil {
			r := m.ReplyToStoryID.String()
			replyToStoryID = &r
		}
		out = append(out, dto.MessageResponse{
			ID:             m.ID.String(),
			ConversationID: m.ConversationID.String(),
			SenderID:       m.SenderID.String(),
			Content:        m.Content,
			MediaURL:       m.MediaURL,
			ReplyToStoryID: replyToStoryID,
			IsRead:         m.IsRead,
			CreatedAt:      m.CreatedAt,
		})
	}
	return out, total, nil
}

func (s *DMService) SendMessage(ctx context.Context, userID, conversationID uuid.UUID, req dto.MessageRequest) (*dto.MessageResponse, error) {
	if userID == uuid.Nil || conversationID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	ok, err := s.repo.ConversationHasParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	content := strings.TrimSpace(deref(req.Content))
	media := strings.TrimSpace(deref(req.MediaURL))
	if content == "" && media == "" {
		return nil, ErrInvalidInput
	}
	msg := &model.Message{ConversationID: conversationID, SenderID: userID, CreatedAt: time.Now()}
	if content != "" {
		msg.Content = &content
	}
	if media != "" {
		msg.MediaURL = &media
	}
	if req.StoryID != nil && strings.TrimSpace(*req.StoryID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.StoryID))
		if err == nil {
			msg.ReplyToStoryID = &parsed
			storyOwnerID, err := s.repo.GetStoryOwner(ctx, parsed)
			if err == nil && storyOwnerID != userID {
				_ = s.repo.CreateNotification(ctx, &model.Notification{RecipientID: storyOwnerID, ActorID: userID, Type: "story_reply", ReferenceID: &msg.ID})
			}
		}
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	var replyToStoryID *string
	if msg.ReplyToStoryID != nil {
		r := msg.ReplyToStoryID.String()
		replyToStoryID = &r
	}
	return &dto.MessageResponse{
		ID:             msg.ID.String(),
		ConversationID: conversationID.String(),
		SenderID:       userID.String(),
		Content:        msg.Content,
		MediaURL:       msg.MediaURL,
		ReplyToStoryID: replyToStoryID,
		IsRead:         msg.IsRead,
		CreatedAt:      msg.CreatedAt,
	}, nil
}

func (s *DMService) ListMessageRequests(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.ConversationResponse, int64, error) {
	if userID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	pg, lm := normalizePagination(page, limit)
	convs, total, err := s.repo.ListMessageRequests(ctx, userID, pg, lm)
	if err != nil {
		return nil, 0, err
	}

	convIDs := make([]uuid.UUID, 0, len(convs))
	for _, c := range convs {
		convIDs = append(convIDs, c.ID)
	}

	parts, err := s.repo.GetConversationParticipants(ctx, convIDs)
	if err != nil {
		return nil, 0, err
	}

	partsMap := make(map[uuid.UUID]*dto.SearchUserItem)
	for _, p := range parts {
		if p.UserID != userID && !p.Conversation.IsGroup {
			partsMap[p.ConversationID] = &dto.SearchUserItem{
				ID: p.User.ID.String(), Name: p.User.Name, Username: p.User.Username, AvatarURL: p.User.AvatarURL,
			}
		}
	}

	out := make([]dto.ConversationResponse, 0, len(convs))
	for _, c := range convs {
		res := dto.ConversationResponse{ID: c.ID.String(), IsGroup: c.IsGroup, CreatedAt: c.CreatedAt}
		if other, ok := partsMap[c.ID]; ok {
			res.OtherUser = other
		}
		out = append(out, res)
	}
	return out, total, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *DMService) UnsendMessage(ctx context.Context, userID, messageID uuid.UUID) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return ErrConversationNotFound
	}
	if msg.SenderID != userID {
		return ErrForbidden
	}
	return s.repo.DeleteMessage(ctx, messageID)
}
