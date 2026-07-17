package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

type SearchService struct{ repo repository.SearchRepository }

func NewSearchService(repo repository.SearchRepository) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) Search(ctx context.Context, viewerID, query string, limit int) (*dto.SearchResponse, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return &dto.SearchResponse{Users: []dto.SearchUserItem{}, Hashtags: []dto.SearchHashtagItem{}}, nil
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	users, err := s.repo.SearchUsers(ctx, viewerID, q, limit)
	if err != nil {
		return nil, err
	}
	hashtags, err := s.repo.SearchHashtags(ctx, viewerID, q, limit)
	if err != nil {
		return nil, err
	}
	return &dto.SearchResponse{Users: buildSearchUsers(users), Hashtags: buildSearchHashtags(hashtags)}, nil
}

func (s *SearchService) GetHashtagPosts(ctx context.Context, tag string, viewerID string, page, limit int) ([]dto.PostResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var vid uuid.UUID
	if viewerID != "" {
		vid, _ = uuid.Parse(viewerID)
	}
	posts, total, err := s.repo.ListPostsByHashtag(ctx, tag, vid, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return buildPosts(posts), total, nil
}

func buildSearchUsers(users []model.User) []dto.SearchUserItem {
	out := make([]dto.SearchUserItem, 0, len(users))
	for _, u := range users {
		out = append(out, dto.SearchUserItem{ID: u.ID.String(), Name: u.Name, Username: u.Username, AvatarURL: u.AvatarURL})
	}
	return out
}

func buildSearchHashtags(items []model.Hashtag) []dto.SearchHashtagItem {
	out := make([]dto.SearchHashtagItem, 0, len(items))
	for _, h := range items {
		out = append(out, dto.SearchHashtagItem{ID: h.ID.String(), Tag: h.Tag})
	}
	return out
}
