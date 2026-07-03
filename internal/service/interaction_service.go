package service

import (
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

type InteractionService struct{ repo repository.InteractionRepository }

func NewInteractionService(repo repository.InteractionRepository) *InteractionService { return &InteractionService{repo: repo} }

func (s *InteractionService) CreateComment(userID, postID uuid.UUID, req dto.CreateCommentRequest) (*dto.CommentResponse, error) {
	if userID == uuid.Nil || postID == uuid.Nil { return nil, ErrInvalidInput }
	content := strings.TrimSpace(req.Content)
	if content == "" || len(content) > 2000 { return nil, ErrInvalidInput }
	exists, err := s.repo.PostExists(postID)
	if err != nil { return nil, err }
	if !exists { return nil, ErrPostNotFound }
	if err := s.ensurePostVisible(userID, postID); err != nil { return nil, err }
	var parentID *uuid.UUID
	if req.ParentCommentID != nil && strings.TrimSpace(*req.ParentCommentID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.ParentCommentID))
		if err != nil { return nil, ErrInvalidInput }
		parentID = &parsed
		parent, err := s.repo.FindCommentByID(parsed)
		if err != nil { return nil, ErrInteractionNotFound }
		if parent.PostID != postID { return nil, ErrInvalidInput }
	}
	comment := &model.Comment{PostID: postID, UserID: userID, Content: content, ParentCommentID: parentID}
	if err := s.repo.CreateComment(comment); err != nil { return nil, err }
	return &dto.CommentResponse{ID: comment.ID.String(), PostID: postID.String(), UserID: userID.String(), Content: comment.Content, IsPinned: comment.IsPinned, CreatedAt: comment.CreatedAt}, nil
}

func (s *InteractionService) DeleteComment(userID, commentID uuid.UUID) error {
	if userID == uuid.Nil || commentID == uuid.Nil { return ErrInvalidInput }
	// Try deleting as comment author first
	err := s.repo.DeleteComment(commentID, userID)
	if err == nil { return nil }
	// If not comment author, check if user is post owner
	postID, postErr := s.repo.GetCommentPostID(commentID)
	if postErr != nil { return ErrInteractionNotFound }
	info, infoErr := s.repo.GetPostOwner(postID)
	if infoErr != nil { return ErrPostNotFound }
	if info.UserID != userID { return ErrForbidden }
	return s.repo.DeleteCommentAsPostOwner(commentID, postID)
}

func (s *InteractionService) LikePost(userID, postID uuid.UUID) (*dto.LikeStatusResponse, error) {
	if userID == uuid.Nil || postID == uuid.Nil { return nil, ErrInvalidInput }
	exists, err := s.repo.PostExists(postID)
	if err != nil { return nil, err }
	if !exists { return nil, ErrPostNotFound }
	if err := s.ensurePostVisible(userID, postID); err != nil { return nil, err }
	if err := s.repo.UpsertLike(&model.Like{UserID: userID, LikeableType: "post", LikeableID: postID}); err != nil { return nil, err }
	return &dto.LikeStatusResponse{TargetID: postID.String(), Liked: true}, nil
}

func (s *InteractionService) UnlikePost(userID, postID uuid.UUID) error {
	if userID == uuid.Nil || postID == uuid.Nil { return ErrInvalidInput }
	return s.repo.DeleteLike(userID, "post", postID)
}

func (s *InteractionService) LikeComment(userID, commentID uuid.UUID) (*dto.LikeStatusResponse, error) {
	if userID == uuid.Nil || commentID == uuid.Nil { return nil, ErrInvalidInput }
	comment, err := s.repo.FindCommentByID(commentID)
	if err != nil { return nil, ErrInteractionNotFound }
	if err := s.ensurePostVisible(userID, comment.PostID); err != nil { return nil, err }
	if err := s.repo.UpsertLike(&model.Like{UserID: userID, LikeableType: "comment", LikeableID: commentID}); err != nil { return nil, err }
	return &dto.LikeStatusResponse{TargetID: commentID.String(), Liked: true}, nil
}

func (s *InteractionService) SavePost(userID, postID uuid.UUID) (*dto.SaveStatusResponse, error) {
	if userID == uuid.Nil || postID == uuid.Nil { return nil, ErrInvalidInput }
	exists, err := s.repo.PostExists(postID)
	if err != nil { return nil, err }
	if !exists { return nil, ErrPostNotFound }
	if err := s.ensurePostVisible(userID, postID); err != nil { return nil, err }
	if err := s.repo.UpsertSavedPost(&model.SavedPost{UserID: userID, PostID: postID, CollectionName: "All"}); err != nil { return nil, err }
	return &dto.SaveStatusResponse{PostID: postID.String(), Saved: true}, nil
}

func (s *InteractionService) UnsavePost(userID, postID uuid.UUID) error {
	if userID == uuid.Nil || postID == uuid.Nil { return ErrInvalidInput }
	return s.repo.DeleteSavedPost(userID, postID)
}

func (s *InteractionService) ensurePostVisible(userID, postID uuid.UUID) error {
	info, err := s.repo.GetPostOwner(postID)
	if err != nil { return ErrPostNotFound }
	if info.UserID == userID { return nil }
	blocked, err := s.repo.IsBlockedEitherDirection(userID, info.UserID)
	if err != nil { return err }
	if blocked { return ErrUserBlocked }
	if info.IsPrivate {
		following, err := s.repo.IsAcceptedFollower(userID, info.UserID)
		if err != nil { return err }
		if !following { return ErrForbidden }
	}
	return nil
}
