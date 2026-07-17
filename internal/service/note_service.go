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

var ErrNoteNotFound = errors.New("note not found")

type NoteService struct{ repo repository.NoteRepository }

func NewNoteService(repo repository.NoteRepository) *NoteService { return &NoteService{repo: repo} }

func (s *NoteService) CreateNote(ctx context.Context, userID uuid.UUID, req dto.CreateNoteRequest) (*dto.NoteResponse, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	content := strings.TrimSpace(req.Content)
	if content == "" || len(content) > 60 {
		return nil, ErrInvalidInput
	}
	var audioID *uuid.UUID
	if req.AudioTrackID != nil && strings.TrimSpace(*req.AudioTrackID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.AudioTrackID))
		if err == nil {
			audioID = &parsed
		}
	}
	note := &model.Note{
		UserID:       userID,
		Content:      content,
		AudioTrackID: audioID,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	if err := s.repo.CreateNote(ctx, note); err != nil {
		return nil, err
	}
	return &dto.NoteResponse{
		ID:           note.ID.String(),
		UserID:       note.UserID.String(),
		Content:      note.Content,
		AudioTrackID: req.AudioTrackID,
		ExpiresAt:    note.ExpiresAt,
		CreatedAt:    note.CreatedAt,
	}, nil
}

func (s *NoteService) GetActiveNotes(ctx context.Context, userID uuid.UUID) ([]dto.NoteResponse, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	notes, err := s.repo.GetActiveNotes(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.NoteResponse, 0, len(notes))
	for _, n := range notes {
		var audio *string
		if n.AudioTrackID != nil {
			a := n.AudioTrackID.String()
			audio = &a
		}
		out = append(out, dto.NoteResponse{
			ID:           n.ID.String(),
			UserID:       n.UserID.String(),
			Username:     n.User.Username,
			AvatarURL:    n.User.AvatarURL,
			Content:      n.Content,
			AudioTrackID: audio,
			ExpiresAt:    n.ExpiresAt,
			CreatedAt:    n.CreatedAt,
		})
	}
	return out, nil
}

func (s *NoteService) DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error {
	if userID == uuid.Nil || noteID == uuid.Nil {
		return ErrInvalidInput
	}
	owns, err := s.repo.OwnsNote(ctx, noteID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}
	return s.repo.DeleteNote(ctx, noteID, userID)
}
