package service

import (
	"strings"

	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

type SearchService struct{ repo repository.SearchRepository }

func NewSearchService(repo repository.SearchRepository) *SearchService { return &SearchService{repo: repo} }

func (s *SearchService) Search(viewerID, query string, limit int) (*dto.SearchResponse, error) {
	q := strings.TrimSpace(query)
	if q == "" { return &dto.SearchResponse{Users: []dto.SearchUserItem{}, Hashtags: []dto.SearchHashtagItem{}}, nil }
	if limit < 1 { limit = 20 }
	if limit > 50 { limit = 50 }
	users, err := s.repo.SearchUsers(viewerID, q, limit)
	if err != nil { return nil, err }
	hashtags, err := s.repo.SearchHashtags(viewerID, q, limit)
	if err != nil { return nil, err }
	return &dto.SearchResponse{Users: buildSearchUsers(users), Hashtags: buildSearchHashtags(hashtags)}, nil
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
