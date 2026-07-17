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

var (
	ErrInteractionNotFound = errors.New("interaction not found")
)

type InteractionService struct {
	repo repository.InteractionRepository
}

func NewInteractionService(repo repository.InteractionRepository) *InteractionService {
	return &InteractionService{repo: repo}
}

func (s *InteractionService) CreateComment(ctx context.Context, userID, postID uuid.UUID, req dto.CreateCommentRequest) (*dto.CommentResponse, error) {
	if userID == uuid.Nil || postID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	content := strings.TrimSpace(req.Content)
	if content == "" || len(content) > 2000 {
		return nil, ErrInvalidInput
	}
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotFound
	}
	if err := s.ensurePostVisible(ctx, userID, postID); err != nil {
		return nil, err
	}
	var parentID *uuid.UUID
	if req.ParentCommentID != nil && strings.TrimSpace(*req.ParentCommentID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.ParentCommentID))
		if err != nil {
			return nil, ErrInvalidInput
		}
		parentID = &parsed
		parent, err := s.repo.FindCommentByID(ctx, parsed)
		if err != nil {
			return nil, ErrInteractionNotFound
		}
		if parent.PostID != postID {
			return nil, ErrInvalidInput
		}
	}
	comment := &model.Comment{PostID: postID, UserID: userID, Content: content, ParentCommentID: parentID}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	ownerID, _ := s.repo.GetPostOwner(ctx, postID)
	if ownerID != nil && ownerID.UserID != userID {
		_ = s.repo.CreateNotification(ctx, &model.Notification{RecipientID: ownerID.UserID, ActorID: userID, Type: "comment", ReferenceID: &postID})
	}

	return &dto.CommentResponse{ID: comment.ID.String(), PostID: postID.String(), UserID: userID.String(), Content: comment.Content, IsPinned: comment.IsPinned, CreatedAt: comment.CreatedAt}, nil
}

func (s *InteractionService) ListComments(ctx context.Context, userID, postID uuid.UUID, page, limit int) ([]dto.CommentResponse, int64, error) {
	if postID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, ErrPostNotFound
	}
	if err := s.ensurePostVisible(ctx, userID, postID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	comments, total, err := s.repo.ListPostComments(ctx, postID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.CommentResponse, 0, len(comments))
	for _, c := range comments {
		var parentID *string
		if c.ParentCommentID != nil {
			p := c.ParentCommentID.String()
			parentID = &p
		}
		out = append(out, dto.CommentResponse{ID: c.ID.String(), PostID: c.PostID.String(), UserID: c.UserID.String(), ParentCommentID: parentID, Content: c.Content, IsPinned: c.IsPinned, CreatedAt: c.CreatedAt})
	}
	return out, total, nil
}

func (s *InteractionService) ListSavedPosts(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.SavedPostResponse, int64, error) {
	if userID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	saved, total, err := s.repo.ListSavedPosts(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.SavedPostResponse, 0, len(saved))
	for _, sv := range saved {
		out = append(out, dto.SavedPostResponse{ID: sv.ID.String(), PostID: sv.PostID.String(), Collection: sv.CollectionName, CreatedAt: sv.CreatedAt, Caption: sv.Post.Caption})
	}
	return out, total, nil
}

func (s *InteractionService) SharePost(ctx context.Context, userID, postID uuid.UUID) (*dto.ShareResponse, error) {
	if postID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotFound
	}
	if err := s.ensurePostVisible(ctx, userID, postID); err != nil {
		return nil, err
	}
	link := "https://twistgram.app/p/" + postID.String()
	return &dto.ShareResponse{Link: link}, nil
}

func (s *InteractionService) DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	if userID == uuid.Nil || commentID == uuid.Nil {
		return ErrInvalidInput
	}
	err := s.repo.DeleteComment(ctx, commentID, userID)
	if err == nil {
		return nil
	}
	postID, postErr := s.repo.GetCommentPostID(ctx, commentID)
	if postErr != nil {
		return ErrInteractionNotFound
	}
	info, infoErr := s.repo.GetPostOwner(ctx, postID)
	if infoErr != nil {
		return ErrPostNotFound
	}
	if info.UserID != userID {
		return ErrForbidden
	}
	return s.repo.DeleteCommentAsPostOwner(ctx, commentID, postID)
}

func (s *InteractionService) LikePost(ctx context.Context, userID, postID uuid.UUID) (*dto.LikeStatusResponse, error) {
	if userID == uuid.Nil || postID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotFound
	}
	if err := s.ensurePostVisible(ctx, userID, postID); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertLike(ctx, &model.Like{UserID: userID, LikeableType: "post", LikeableID: postID}); err != nil {
		return nil, err
	}

	ownerID, _ := s.repo.GetPostOwner(ctx, postID)
	if ownerID != nil && ownerID.UserID != userID {
		_ = s.repo.CreateNotification(ctx, &model.Notification{RecipientID: ownerID.UserID, ActorID: userID, Type: "like", ReferenceID: &postID})
	}

	return &dto.LikeStatusResponse{TargetID: postID.String(), Liked: true}, nil
}

func (s *InteractionService) UnlikePost(ctx context.Context, userID, postID uuid.UUID) error {
	if userID == uuid.Nil || postID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.DeleteLike(ctx, userID, "post", postID)
}

func (s *InteractionService) LikeComment(ctx context.Context, userID, commentID uuid.UUID) (*dto.LikeStatusResponse, error) {
	if userID == uuid.Nil || commentID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	comment, err := s.repo.FindCommentByID(ctx, commentID)
	if err != nil {
		return nil, ErrInteractionNotFound
	}
	if err := s.ensurePostVisible(ctx, userID, comment.PostID); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertLike(ctx, &model.Like{UserID: userID, LikeableType: "comment", LikeableID: commentID}); err != nil {
		return nil, err
	}
	return &dto.LikeStatusResponse{TargetID: commentID.String(), Liked: true}, nil
}

func (s *InteractionService) SavePost(ctx context.Context, userID, postID uuid.UUID) (*dto.SaveStatusResponse, error) {
	if userID == uuid.Nil || postID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotFound
	}
	if err := s.ensurePostVisible(ctx, userID, postID); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertSavedPost(ctx, &model.SavedPost{UserID: userID, PostID: postID, CollectionName: "All"}); err != nil {
		return nil, err
	}
	return &dto.SaveStatusResponse{PostID: postID.String(), Saved: true}, nil
}

func (s *InteractionService) UnsavePost(ctx context.Context, userID, postID uuid.UUID) error {
	if userID == uuid.Nil || postID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.DeleteSavedPost(ctx, userID, postID)
}

func (s *InteractionService) ensurePostVisible(ctx context.Context, userID, postID uuid.UUID) error {
	info, err := s.repo.GetPostOwner(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}
	if info.UserID == userID {
		return nil
	}
	blocked, err := s.repo.IsBlockedEitherDirection(ctx, userID, info.UserID)
	if err != nil {
		return err
	}
	if blocked {
		return ErrUserBlocked
	}
	if info.IsPrivate {
		following, err := s.repo.IsAcceptedFollower(ctx, userID, info.UserID)
		if err != nil {
			return err
		}
		if !following {
			return ErrForbidden
		}
	}
	return nil
}
