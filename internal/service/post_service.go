package service

import (
	"context"
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
var hashtagRegex = regexp.MustCompile("#[^\\s!@#$%^&*()=+.//,\\[{\\]};:'\"?><]+")

type PostService struct{ repo repository.PostRepository }

func NewPostService(repo repository.PostRepository) *PostService { return &PostService{repo: repo} }

func (s *PostService) CreatePost(ctx context.Context, userID uuid.UUID, req dto.CreatePostRequest) (*dto.PostResponse, error) {
	if userID == uuid.Nil || len(req.Media) == 0 {
		return nil, ErrInvalidInput
	}
	if err := validateCreatePostRequest(req); err != nil {
		return nil, err
	}

	post := &model.Post{UserID: userID, Caption: cleanOptional(req.Caption), Location: cleanOptional(req.Location)}
	if req.CommentsDisabled != nil {
		post.CommentsDisabled = *req.CommentsDisabled
	}

	media := make([]model.PostMedia, 0, len(req.Media))
	for _, m := range req.Media {
		media = append(media, model.PostMedia{MediaURL: strings.TrimSpace(m.MediaURL), MediaType: strings.TrimSpace(strings.ToLower(m.MediaType)), OrderIndex: m.OrderIndex, MusicTrackURL: cleanOptional(m.MusicTrackURL)})
	}

	tags := make([]model.PostTag, 0)
	for _, idStr := range req.TaggedUserIDs {
		parsed, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil || parsed == userID {
			continue
		}
		tags = append(tags, model.PostTag{TaggedUserID: parsed})
	}

	var hashtags []model.Hashtag
	if post.Caption != nil {
		matches := hashtagRegex.FindAllString(*post.Caption, -1)
		for _, match := range matches {
			tag := strings.ToLower(strings.TrimPrefix(match, "#"))
			if tag != "" {
				hashtags = append(hashtags, model.Hashtag{Tag: tag})
			}
		}
	}

	if err := s.repo.CreatePostWithMediaAndTags(ctx, post, media, tags, hashtags); err != nil {
		return nil, err
	}

	for _, tag := range tags {
		_ = s.repo.CreateNotification(ctx, &model.Notification{RecipientID: tag.TaggedUserID, ActorID: userID, Type: "mention", ReferenceID: &post.ID})
	}

	// Re-fetch post with relations
	p, _ := s.repo.GetPostWithMedia(ctx, post.ID)
	if p == nil { p = post }

	res := buildPosts([]model.Post{*p})
	return &res[0], nil
}

func validateCreatePostRequest(req dto.CreatePostRequest) error {
	if req.Caption != nil && len(strings.TrimSpace(*req.Caption)) > 2200 {
		return ErrInvalidInput
	}
	if req.Location != nil && len(strings.TrimSpace(*req.Location)) > 255 {
		return ErrInvalidInput
	}
	for _, m := range req.Media {
		if strings.TrimSpace(m.MediaURL) == "" {
			return ErrInvalidInput
		}
		t := strings.TrimSpace(strings.ToLower(m.MediaType))
		if t != "image" && t != "video" {
			return ErrInvalidInput
		}
	}
	return nil
}

func (s *PostService) Feed(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) {
	posts, total, err := s.repo.ListFeed(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		fbPosts, fbTotal, fbErr := s.repo.ListGlobalFeed(ctx, userID, limit)
		if fbErr != nil {
			return nil, 0, fbErr
		}
		return buildPosts(fbPosts), fbTotal, nil
	}
	return buildPosts(posts), total, nil
}

func (s *PostService) MyPosts(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) {
	posts, total, err := s.repo.ListUserPosts(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return buildPosts(posts), total, nil
}

func (s *PostService) UserPosts(ctx context.Context, viewerID, targetUserID uuid.UUID, page, limit int) ([]dto.PostResponse, int64, error) {
	if targetUserID == uuid.Nil {
		return nil, 0, ErrPostNotFound
	}

	if viewerID == targetUserID {
		posts, total, err := s.repo.ListUserPosts(ctx, targetUserID, page, limit)
		if err != nil {
			return nil, 0, err
		}
		return buildPosts(posts), total, nil
	}

	blocked, err := s.repo.IsBlockedEitherDirection(ctx, viewerID, targetUserID)
	if err != nil {
		return nil, 0, err
	}
	if blocked {
		return nil, 0, ErrForbidden
	}

	isPrivate, err := s.repo.IsUserPrivate(ctx, targetUserID)
	if err != nil {
		return nil, 0, err
	}

	if isPrivate {
		isFollower, err := s.repo.IsAcceptedFollower(ctx, viewerID, targetUserID)
		if err != nil {
			return nil, 0, err
		}
		if !isFollower {
			return nil, 0, ErrForbidden
		}
	}

	posts, total, err := s.repo.ListUserPosts(ctx, targetUserID, page, limit)
	if err != nil {
		return nil, 0, err
	}

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

func (s *PostService) Archive(ctx context.Context, userID, postID uuid.UUID, archived bool) error {
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPostNotFound
	}
	owns, err := s.repo.OwnsPost(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}
	return s.repo.SetArchived(ctx, postID, userID, archived)
}

func (s *PostService) EditCaption(ctx context.Context, userID, postID uuid.UUID, req dto.UpdatePostRequest) error {
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPostNotFound
	}
	owns, err := s.repo.OwnsPost(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}

	if req.Caption != nil && len(strings.TrimSpace(*req.Caption)) > 2200 {
		return ErrInvalidInput
	}
	return s.repo.UpdateCaption(ctx, postID, userID, cleanOptional(req.Caption))
}

func (s *PostService) Delete(ctx context.Context, userID, postID uuid.UUID) error {
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPostNotFound
	}
	owns, err := s.repo.OwnsPost(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return ErrForbidden
	}
	return s.repo.DeletePost(ctx, postID, userID)
}

func (s *PostService) RemoveTag(ctx context.Context, userID, postID, taggedUserID uuid.UUID) error {
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPostNotFound
	}
	owns, err := s.repo.OwnsPost(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !owns && userID != taggedUserID {
		return ErrForbidden
	}
	return s.repo.DeletePostTag(ctx, postID, taggedUserID)
}

func (s *PostService) GetByID(ctx context.Context, userID, postID uuid.UUID) (*dto.PostResponse, error) {
	if postID == uuid.Nil {
		return nil, ErrPostNotFound
	}
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	if post.IsArchived && post.UserID != userID {
		return nil, ErrPostNotFound
	}

	res := buildPosts([]model.Post{*post})
	return &res[0], nil
}

func buildPosts(posts []model.Post) []dto.PostResponse {
	out := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		resp := dto.PostResponse{
			PostID:           p.ID.String(),
			PostType:         "IMAGE", // default fallback
			User: dto.PostUserResponse{
				UserID:   p.UserID.String(),
				Username: p.User.Username,
			},
			Caption:          p.Caption,
			Location:         p.Location,
			IsArchived:       p.IsArchived,
			CommentsDisabled: p.CommentsDisabled,
			CreatedAt:        p.CreatedAt,
			Metrics: dto.PostMetricsResponse{
				// MVP Phase: Set defaults; fully query these dynamically via repository
				LikesCount:    0,
				CommentsCount: 0,
				HasLiked:      false,
				HasSaved:      false,
			},
		}

		if p.User.AvatarURL != nil {
			resp.User.AvatarURL = p.User.AvatarURL
		}
		// Try to pull is_verified if we fetch it (Phase 3, but default false)
		
		if len(p.Media) > 0 {
			if len(p.Media) == 1 {
				resp.PostType = strings.ToUpper(p.Media[0].MediaType)
				url := p.Media[0].MediaURL
				resp.MediaURL = &url
			} else {
				resp.PostType = "CAROUSEL"
				resp.MediaItems = make([]dto.PostMediaResponse, 0, len(p.Media))
				for _, m := range p.Media {
					resp.MediaItems = append(resp.MediaItems, dto.PostMediaResponse{
						MediaID: m.ID.String(),
						Type:    strings.ToUpper(m.MediaType),
						URL:     m.MediaURL,
					})
				}
			}
		}
		out = append(out, resp)
	}
	return out
}

func cleanOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}



