package service

import (
	"strings"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/repository"
)

type SearchHistoryService struct {
	repo repository.SearchHistoryRepository
}

func NewSearchHistoryService(repo repository.SearchHistoryRepository) *SearchHistoryService {
	return &SearchHistoryService{repo: repo}
}

func (s *SearchHistoryService) ListHistory(userID uuid.UUID, limit int) ([]dto.SearchHistoryItem, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	items, err := s.repo.ListHistory(userID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SearchHistoryItem, 0, len(items))
	for _, item := range items {
		out = append(out, dto.SearchHistoryItem{
			ID:        item.ID.String(),
			Query:     item.Query,
			QueryType: item.QueryType,
			CreatedAt: item.CreatedAt,
		})
	}
	return out, nil
}

func (s *SearchHistoryService) SaveSearch(userID uuid.UUID, query, queryType string) error {
	if userID == uuid.Nil {
		return ErrInvalidInput
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return ErrInvalidInput
	}
	queryType = strings.TrimSpace(strings.ToLower(queryType))
	if queryType != "user" && queryType != "hashtag" {
		queryType = "user"
	}
	return s.repo.SaveSearch(userID, q, queryType)
}

func (s *SearchHistoryService) DeleteHistoryItem(userID, itemID uuid.UUID) error {
	if userID == uuid.Nil || itemID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.DeleteHistoryItem(itemID, userID)
}

func (s *SearchHistoryService) DeleteAllHistory(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.DeleteAllHistory(userID)
}
