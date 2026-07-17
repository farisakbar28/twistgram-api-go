package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

var ErrHighlightNotFound = errors.New("highlight not found")

type HighlightService struct {
	repo repository.HighlightRepository
}

func NewHighlightService(repo repository.HighlightRepository) *HighlightService {
	return &HighlightService{repo: repo}
}

func (s *HighlightService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateHighlightRequest) (*dto.HighlightResponse, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	title := strings.TrimSpace(req.Title)
	if title == "" || len(title) > 100 {
		return nil, ErrInvalidInput
	}
	highlight := &model.Highlight{
		UserID: userID,
		Title:  title,
	}
	if err := s.repo.CreateHighlight(ctx, highlight); err != nil {
		return nil, err
	}
	return buildHighlightResponse(highlight, nil), nil
}

func (s *HighlightService) List(ctx context.Context, userID uuid.UUID) ([]dto.HighlightResponse, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	highlights, err := s.repo.ListHighlightsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.HighlightResponse, 0, len(highlights))
	for _, h := range highlights {
		stories, _ := s.repo.ListHighlightStories(ctx, h.ID)
		resp := buildHighlightResponse(&h, stories)
		out = append(out, *resp)
	}
	return out, nil
}

func (s *HighlightService) Update(ctx context.Context, userID, highlightID uuid.UUID, req dto.UpdateHighlightRequest) error {
	if userID == uuid.Nil || highlightID == uuid.Nil {
		return ErrInvalidInput
	}
	owns, err := s.repo.OwnsHighlight(ctx, highlightID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}
	highlight, err := s.repo.GetHighlightByID(ctx, highlightID)
	if err != nil {
		return ErrHighlightNotFound
	}
	title := strings.TrimSpace(req.Title)
	if title == "" || len(title) > 100 {
		return ErrInvalidInput
	}
	highlight.Title = title
	return s.repo.UpdateHighlight(ctx, highlight)
}

func (s *HighlightService) Delete(ctx context.Context, userID, highlightID uuid.UUID) error {
	if userID == uuid.Nil || highlightID == uuid.Nil {
		return ErrInvalidInput
	}
	owns, err := s.repo.OwnsHighlight(ctx, highlightID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}
	return s.repo.DeleteHighlight(ctx, highlightID)
}

func (s *HighlightService) AddStory(ctx context.Context, userID, highlightID uuid.UUID, req dto.AddStoryToHighlightRequest) error {
	if userID == uuid.Nil || highlightID == uuid.Nil {
		return ErrInvalidInput
	}
	owns, err := s.repo.OwnsHighlight(ctx, highlightID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}
	storyID, err := uuid.Parse(strings.TrimSpace(req.StoryID))
	if err != nil {
		return ErrInvalidInput
	}
	return s.repo.AddStoryToHighlight(ctx, highlightID, storyID)
}

func (s *HighlightService) RemoveStory(ctx context.Context, userID, highlightID, storyID uuid.UUID) error {
	if userID == uuid.Nil || highlightID == uuid.Nil || storyID == uuid.Nil {
		return ErrInvalidInput
	}
	owns, err := s.repo.OwnsHighlight(ctx, highlightID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}
	return s.repo.RemoveStoryFromHighlight(ctx, highlightID, storyID)
}

func buildHighlightResponse(h *model.Highlight, stories []model.Story) *dto.HighlightResponse {
	resp := &dto.HighlightResponse{
		ID:        h.ID.String(),
		Title:     h.Title,
		CreatedAt: h.CreatedAt,
	}
	if len(stories) > 0 {
		resp.Stories = make([]dto.HighlightStoryResponse, 0, len(stories))
		for _, s := range stories {
			resp.Stories = append(resp.Stories, dto.HighlightStoryResponse{
				ID:        s.ID.String(),
				StoryID:   s.ID.String(),
				MediaURL:  s.MediaURL,
				MediaType: s.MediaType,
				CreatedAt: s.CreatedAt,
			})
		}
	}
	return resp
}
