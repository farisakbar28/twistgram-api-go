package service

import (
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

func (s *DMService) StartConversation(userID, targetID uuid.UUID) (*dto.ConversationResponse, error) {
	if userID == uuid.Nil || targetID == uuid.Nil || userID == targetID { return nil, ErrInvalidInput }
	blocked, err := s.repo.UserBlocked(userID, targetID)
	if err != nil { return nil, err }
	if blocked { return nil, ErrForbidden }
	if existing, err := s.repo.FindConversationBetween(userID, targetID); err == nil && existing != nil {
		return &dto.ConversationResponse{ID: existing.ID.String(), IsGroup: existing.IsGroup, CreatedAt: existing.CreatedAt}, nil
	}
	conv := &model.Conversation{}
	if err := s.repo.CreateConversation(conv); err != nil { return nil, err }
	if err := s.repo.AddParticipant(&model.ConversationParticipant{ConversationID: conv.ID, UserID: userID}); err != nil { return nil, err }
	if err := s.repo.AddParticipant(&model.ConversationParticipant{ConversationID: conv.ID, UserID: targetID}); err != nil { return nil, err }
	return &dto.ConversationResponse{ID: conv.ID.String(), IsGroup: conv.IsGroup, CreatedAt: conv.CreatedAt}, nil
}

func (s *DMService) ListConversations(userID uuid.UUID, page, limit int) ([]dto.ConversationResponse, int64, error) {
	if userID == uuid.Nil { return nil, 0, ErrInvalidInput }
	pg, lm := normalizePagination(page, limit)
	convs, total, err := s.repo.ListConversations(userID, pg, lm)
	if err != nil { return nil, 0, err }
	
	convIDs := make([]uuid.UUID, 0, len(convs))
	for _, c := range convs { convIDs = append(convIDs, c.ID) }
	
	parts, err := s.repo.GetConversationParticipants(convIDs)
	if err != nil { return nil, 0, err }
	
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
		if other, ok := partsMap[c.ID]; ok { res.OtherUser = other }
		out = append(out, res)
	}
	return out, total, nil
}

func (s *DMService) ListMessages(userID, conversationID uuid.UUID, page, limit int) ([]dto.MessageResponse, int64, error) {
	if userID == uuid.Nil || conversationID == uuid.Nil { return nil, 0, ErrInvalidInput }
	ok, err := s.repo.ConversationHasParticipant(conversationID, userID)
	if err != nil { return nil, 0, err }
	if !ok { return nil, 0, ErrForbidden }
	pg, lm := normalizePagination(page, limit)
	msgs, total, err := s.repo.ListMessages(conversationID, pg, lm)
	if err != nil { return nil, 0, err }
	out := make([]dto.MessageResponse, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, dto.MessageResponse{ID: m.ID.String(), ConversationID: m.ConversationID.String(), SenderID: m.SenderID.String(), Content: m.Content, MediaURL: m.MediaURL, IsRead: m.IsRead, CreatedAt: m.CreatedAt})
	}
	return out, total, nil
}

func (s *DMService) SendMessage(userID, conversationID uuid.UUID, req dto.MessageRequest) (*dto.MessageResponse, error) {
	if userID == uuid.Nil || conversationID == uuid.Nil { return nil, ErrInvalidInput }
	ok, err := s.repo.ConversationHasParticipant(conversationID, userID)
	if err != nil { return nil, err }
	if !ok { return nil, ErrForbidden }
	content := strings.TrimSpace(deref(req.Content))
	media := strings.TrimSpace(deref(req.MediaURL))
	if content == "" && media == "" { return nil, ErrInvalidInput }
	msg := &model.Message{ConversationID: conversationID, SenderID: userID, CreatedAt: time.Now()}
	if content != "" { msg.Content = &content }
	if media != "" { msg.MediaURL = &media }
	if req.StoryID != nil && strings.TrimSpace(*req.StoryID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.StoryID))
		if err == nil {
			msg.ReplyToStoryID = &parsed
			// CNT-05/§5.5: Notifikasi ke pemilik story saat ada reply
			storyOwnerID, err := s.repo.GetStoryOwner(parsed)
			if err == nil && storyOwnerID != userID {
				_ = s.repo.CreateNotification(&model.Notification{RecipientID: storyOwnerID, ActorID: userID, Type: "story_reply", ReferenceID: &msg.ID})
			}
		}
	}
	if err := s.repo.CreateMessage(msg); err != nil { return nil, err }
	return &dto.MessageResponse{ID: msg.ID.String(), ConversationID: conversationID.String(), SenderID: userID.String(), Content: msg.Content, MediaURL: msg.MediaURL, IsRead: msg.IsRead, CreatedAt: msg.CreatedAt}, nil
}

// MSG-02: List conversations where the other user has private account
// and current user is NOT an accepted follower — these are "message requests"
func (s *DMService) ListMessageRequests(userID uuid.UUID, page, limit int) ([]dto.ConversationResponse, int64, error) {
	if userID == uuid.Nil { return nil, 0, ErrInvalidInput }
	pg, lm := normalizePagination(page, limit)
	convs, total, err := s.repo.ListMessageRequests(userID, pg, lm)
	if err != nil { return nil, 0, err }

	convIDs := make([]uuid.UUID, 0, len(convs))
	for _, c := range convs { convIDs = append(convIDs, c.ID) }

	parts, err := s.repo.GetConversationParticipants(convIDs)
	if err != nil { return nil, 0, err }

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
		if other, ok := partsMap[c.ID]; ok { res.OtherUser = other }
		out = append(out, res)
	}
	return out, total, nil
}

func deref(s *string) string { if s == nil { return "" }; return *s }
