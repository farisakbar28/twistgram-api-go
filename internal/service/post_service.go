package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

var ErrPostNotFound = errors.New("post not found")
var ErrForbidden = errors.New("forbidden")

type PostService struct{ repo repository.PostRepository }

func NewPostService(repo repository.PostRepository) *PostService { return &PostService{repo: repo} }

func (s *PostService) CreatePost(userID uuid.UUID, req dto.CreatePostRequest) (*dto.PostResponse, error) {
	if userID == uuid.Nil || len(req.Media) == 0 { return nil, ErrInvalidInput }
	if err := validateCreatePostRequest(req); err != nil { return nil, err }
	post := &model.Post{UserID: userID, Caption: cleanOptional(req.Caption), Location: cleanOptional(req.Location)}
	if req.CommentsDisabled != nil { post.CommentsDisabled = *req.CommentsDisabled }
	media := make([]model.PostMedia, 0, len(req.Media))
	for _, m := range req.Media { media = append(media, model.PostMedia{MediaURL: strings.TrimSpace(m.MediaURL), MediaType: strings.TrimSpace(strings.ToLower(m.MediaType)), OrderIndex: m.OrderIndex, MusicTrackURL: cleanOptional(m.MusicTrackURL)}) }
	if err := s.repo.CreatePostWithMedia(post, media); err != nil { return nil, err }
	return &dto.PostResponse{ID: post.ID.String(), UserID: userID.String(), Caption: post.Caption, Location: post.Location, IsArchived: post.IsArchived, CommentsDisabled: post.CommentsDisabled, CreatedAt: post.CreatedAt}, nil
}

func validateCreatePostRequest(req dto.CreatePostRequest) error {
	if req.Caption != nil && len(strings.TrimSpace(*req.Caption)) > 2200 { return ErrInvalidInput }
	if req.Location != nil && len(strings.TrimSpace(*req.Location)) > 255 { return ErrInvalidInput }
	for _, m := range req.Media {
		if strings.TrimSpace(m.MediaURL) == "" { return ErrInvalidInput }
		t := strings.TrimSpace(strings.ToLower(m.MediaType))
		if t != "image" && t != "video" { return ErrInvalidInput }
	}
	return nil
}

func (s *PostService) Feed(userID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) { posts, total, err := s.repo.ListFeed(userID, page, limit); if err != nil { return nil, 0, err }; return buildPosts(posts), total, nil }
func (s *PostService) MyPosts(userID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) { posts, total, err := s.repo.ListUserPosts(userID, page, limit); if err != nil { return nil, 0, err }; return buildPosts(posts), total, nil }
func (s *PostService) Archive(userID, postID uuid.UUID, archived bool) error {
	exists, err := s.repo.PostExists(postID)
	if err != nil { return err }
	if !exists { return ErrPostNotFound }
	owns, err := s.repo.OwnsPost(postID, userID)
	if err != nil { return err }
	if !owns { return ErrForbidden }
	return s.repo.SetArchived(postID, userID, archived)
}
func (s *PostService) Delete(userID, postID uuid.UUID) error {
	exists, err := s.repo.PostExists(postID)
	if err != nil { return err }
	if !exists { return ErrPostNotFound }
	owns, err := s.repo.OwnsPost(postID, userID)
	if err != nil { return err }
	if !owns { return ErrForbidden }
	return s.repo.DeletePost(postID, userID)
}

func buildPosts(posts []model.Post) []dto.PostResponse { out:=make([]dto.PostResponse,0,len(posts)); for _, p := range posts { out = append(out, dto.PostResponse{ID:p.ID.String(), UserID:p.UserID.String(), Caption:p.Caption, Location:p.Location, IsArchived:p.IsArchived, CommentsDisabled:p.CommentsDisabled, CreatedAt:p.CreatedAt})}; return out }

func cleanOptional(value *string) *string { if value == nil { return nil }; trimmed := strings.TrimSpace(*value); if trimmed == "" { return nil }; return &trimmed }
