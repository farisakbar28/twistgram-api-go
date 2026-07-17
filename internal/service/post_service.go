package service

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

var ErrPostNotFound = errors.New("post not found")
var ErrForbidden = errors.New("forbidden")
var hashtagRegex = regexp.MustCompile(`#[^\s!@#$%^&*()=+.\/,\[{\]};:'"?><]+`)

type PostService struct{ repo repository.PostRepository }

func NewPostService(repo repository.PostRepository) *PostService { return &PostService{repo: repo} }

func (s *PostService) CreatePost(userID uuid.UUID, req dto.CreatePostRequest) (*dto.PostResponse, error) {
	if userID == uuid.Nil || len(req.Media) == 0 { return nil, ErrInvalidInput }
	if err := validateCreatePostRequest(req); err != nil { return nil, err }
	
	post := &model.Post{UserID: userID, Caption: cleanOptional(req.Caption), Location: cleanOptional(req.Location)}
	if req.CommentsDisabled != nil { post.CommentsDisabled = *req.CommentsDisabled }
	
	media := make([]model.PostMedia, 0, len(req.Media))
	for _, m := range req.Media { media = append(media, model.PostMedia{MediaURL: strings.TrimSpace(m.MediaURL), MediaType: strings.TrimSpace(strings.ToLower(m.MediaType)), OrderIndex: m.OrderIndex, MusicTrackURL: cleanOptional(m.MusicTrackURL)}) }
	
	tags := make([]model.PostTag, 0)
	for _, idStr := range req.TaggedUserIDs {
		parsed, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil || parsed == userID { continue }
		tags = append(tags, model.PostTag{TaggedUserID: parsed})
	}
	
	var hashtags []model.Hashtag
	if post.Caption != nil {
		matches := hashtagRegex.FindAllString(*post.Caption, -1)
		for _, match := range matches {
			tag := strings.ToLower(strings.TrimPrefix(match, "#"))
			if tag != "" { hashtags = append(hashtags, model.Hashtag{Tag: tag}) }
		}
	}
	
	if err := s.repo.CreatePostWithMediaAndTags(post, media, tags, hashtags); err != nil { return nil, err }
	
	// Send mention notifications
	for _, tag := range tags {
		_ = s.repo.CreateNotification(&model.Notification{RecipientID: tag.TaggedUserID, ActorID: userID, Type: "mention", ReferenceID: &post.ID})
	}
	
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

func (s *PostService) Feed(userID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) {
	posts, total, err := s.repo.ListFeed(userID, page, limit)
	if err != nil { return nil, 0, err }
	if total == 0 {
		// Fallback to global explore if no following posts
		fbPosts, fbTotal, fbErr := s.repo.ListGlobalFeed(limit)
		if fbErr != nil { return nil, 0, fbErr }
		return buildPosts(fbPosts), fbTotal, nil
	}
	return buildPosts(posts), total, nil
}
func (s *PostService) MyPosts(userID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) { posts, total, err := s.repo.ListUserPosts(userID, page, limit); if err != nil { return nil, 0, err }; return buildPosts(posts), total, nil }

// UserPosts returns posts for a specific user with privacy and block checks.
func (s *PostService) UserPosts(viewerID, targetUserID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) {
	if targetUserID == uuid.Nil { return nil, 0, ErrPostNotFound }

	// If viewing own posts, show all non-archived
	if viewerID == targetUserID {
		posts, total, err := s.repo.ListUserPosts(targetUserID, page, limit)
		if err != nil { return nil, 0, err }
		return buildPosts(posts), total, nil
	}

	// Check if blocked (either direction)
	blocked, err := s.repo.IsBlockedEitherDirection(viewerID, targetUserID)
	if err != nil { return nil, 0, err }
	if blocked { return nil, 0, ErrForbidden }

	// Check if target is private
	isPrivate, err := s.repo.IsUserPrivate(targetUserID)
	if err != nil { return nil, 0, err }

	// If private, check if viewer is accepted follower
	if isPrivate {
		isFollower, err := s.repo.IsAcceptedFollower(viewerID, targetUserID)
		if err != nil { return nil, 0, err }
		if !isFollower { return nil, 0, ErrForbidden }
	}

	posts, total, err := s.repo.ListUserPosts(targetUserID, page, limit)
	if err != nil { return nil, 0, err }

	// Filter out archived posts for non-owners
	if len(posts) > 0 {
		filtered := make([]model.Post, 0, len(posts))
		for _, p := range posts {
			if !p.IsArchived {
				filtered = append(filtered, p)
			}
		}
		return buildPosts(filtered), int64(len(filtered)), nil
	}

	return buildPosts(posts), total, nil
}
func (s *PostService) Archive(userID, postID uuid.UUID, archived bool) error {
	exists, err := s.repo.PostExists(postID)
	if err != nil { return err }
	if !exists { return ErrPostNotFound }
	owns, err := s.repo.OwnsPost(postID, userID)
	if err != nil { return err }
	if !owns { return ErrForbidden }
	return s.repo.SetArchived(postID, userID, archived)
}
func (s *PostService) EditCaption(userID, postID uuid.UUID, req dto.UpdatePostRequest) error {
	exists, err := s.repo.PostExists(postID)
	if err != nil { return err }
	if !exists { return ErrPostNotFound }
	owns, err := s.repo.OwnsPost(postID, userID)
	if err != nil { return err }
	if !owns { return ErrForbidden }
	
	if req.Caption != nil && len(strings.TrimSpace(*req.Caption)) > 2200 { return ErrInvalidInput }
	return s.repo.UpdateCaption(postID, userID, cleanOptional(req.Caption))
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

func (s *PostService) RemoveTag(userID, postID, taggedUserID uuid.UUID) error {
	exists, err := s.repo.PostExists(postID)
	if err != nil { return err }
	if !exists { return ErrPostNotFound }
	owns, err := s.repo.OwnsPost(postID, userID)
	if err != nil { return err }
	if !owns { return ErrForbidden }
	return s.repo.DeletePostTag(postID, taggedUserID)
}

// GetByID returns a single post with full detail (media, tags, hashtags, user).
func (s *PostService) GetByID(userID, postID uuid.UUID) (*dto.PostResponse, error) {
	if postID == uuid.Nil { return nil, ErrPostNotFound }
	post, err := s.repo.GetPostByID(postID)
	if err != nil { return nil, ErrPostNotFound }

	// CNT-03: archived posts visible only to owner
	if post.IsArchived && post.UserID != userID {
		return nil, ErrPostNotFound
	}

	resp := &dto.PostResponse{
		ID:               post.ID.String(),
		UserID:           post.UserID.String(),
		Caption:          post.Caption,
		Location:         post.Location,
		IsArchived:       post.IsArchived,
		CommentsDisabled: post.CommentsDisabled,
		CreatedAt:        post.CreatedAt,
	}
	if len(post.Media) > 0 {
		resp.Media = make([]dto.PostMediaResponse, 0, len(post.Media))
		for _, m := range post.Media {
			resp.Media = append(resp.Media, dto.PostMediaResponse{
				MediaURL:      m.MediaURL,
				MediaType:     m.MediaType,
				OrderIndex:    m.OrderIndex,
				MusicTrackURL: m.MusicTrackURL,
			})
		}
	}
	return resp, nil
}

func buildPosts(posts []model.Post) []dto.PostResponse {
	out := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		resp := dto.PostResponse{
			ID:               p.ID.String(),
			UserID:           p.UserID.String(),
			Caption:          p.Caption,
			Location:         p.Location,
			IsArchived:       p.IsArchived,
			CommentsDisabled: p.CommentsDisabled,
			CreatedAt:        p.CreatedAt,
		}
		if len(p.Media) > 0 {
			resp.Media = make([]dto.PostMediaResponse, 0, len(p.Media))
			for _, m := range p.Media {
				resp.Media = append(resp.Media, dto.PostMediaResponse{
					MediaURL:      m.MediaURL,
					MediaType:     m.MediaType,
					OrderIndex:    m.OrderIndex,
					MusicTrackURL: m.MusicTrackURL,
				})
			}
		}
		out = append(out, resp)
	}
	return out
}

func cleanOptional(value *string) *string { if value == nil { return nil }; trimmed := strings.TrimSpace(*value); if trimmed == "" { return nil }; return &trimmed }
