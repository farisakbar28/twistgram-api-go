package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/pkg/response"
)

var (
	ErrSelfAction     = errors.New("cannot perform action on yourself")
	ErrBlocked        = errors.New("blocked relationship")
	ErrFollowNotFound = errors.New("follow request not found")
	ErrTargetNotFound = errors.New("target not found")
	ErrInvalidTarget  = errors.New("invalid target")
	ErrInvalidReason  = errors.New("invalid report reason")
	ErrUserBlocked    = errors.New("user blocked")
)

type SocialService struct {
	repo repository.SocialRepository
}

func NewSocialService(repo repository.SocialRepository) *SocialService {
	return &SocialService{repo: repo}
}

func (s *SocialService) Follow(ctx context.Context, viewerID, targetID uuid.UUID) (*dto.FollowStatusResponse, error) {
	if viewerID == targetID {
		return nil, ErrSelfAction
	}
	target, err := s.repo.FindUserByID(ctx, targetID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	blocked, err := s.repo.IsBlockedEitherDirection(ctx, viewerID, targetID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrBlocked
	}
	status := "accepted"
	if target.IsPrivate {
		status = "pending"
	}
	if err := s.repo.UpsertFollow(ctx, &model.Follow{FollowerID: viewerID, FollowingID: targetID, Status: status}); err != nil {
		return nil, err
	}
	notifType := "follow"
	if status == "pending" {
		notifType = "follow_request"
	}
	_ = s.repo.CreateNotification(ctx, &model.Notification{RecipientID: targetID, ActorID: viewerID, Type: notifType})
	return &dto.FollowStatusResponse{TargetID: targetID.String(), Status: status}, nil
}

func (s *SocialService) Unfollow(ctx context.Context, viewerID, targetID uuid.UUID) error {
	if viewerID == targetID {
		return ErrSelfAction
	}
	if err := s.ensureUserExists(ctx, targetID); err != nil {
		return err
	}
	return s.repo.DeleteFollow(ctx, viewerID, targetID)
}

func (s *SocialService) RemoveFollower(ctx context.Context, viewerID, followerID uuid.UUID) error {
	if viewerID == followerID {
		return ErrSelfAction
	}
	if err := s.ensureUserExists(ctx, followerID); err != nil {
		return err
	}
	return s.repo.DeleteFollow(ctx, followerID, viewerID)
}

func (s *SocialService) ListFollowers(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.SocialUserResponse, *response.Meta, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, nil, err
	}
	page, limit = normalizePagination(page, limit)
	users, total, err := s.repo.ListFollowers(ctx, userID, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return buildSocialUsers(users), buildMeta(page, limit, total), nil
}

func (s *SocialService) ListFollowing(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.SocialUserResponse, *response.Meta, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, nil, err
	}
	page, limit = normalizePagination(page, limit)
	users, total, err := s.repo.ListFollowing(ctx, userID, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return buildSocialUsers(users), buildMeta(page, limit, total), nil
}

func (s *SocialService) ListIncomingFollowRequests(ctx context.Context, viewerID uuid.UUID, page, limit int) ([]dto.FollowRequestResponse, *response.Meta, error) {
	page, limit = normalizePagination(page, limit)
	follows, total, err := s.repo.ListIncomingFollowRequests(ctx, viewerID, page, limit)
	if err != nil {
		return nil, nil, err
	}
	items := make([]dto.FollowRequestResponse, 0, len(follows))
	for _, follow := range follows {
		items = append(items, dto.FollowRequestResponse{ID: follow.ID.String(), Requester: buildSocialUser(follow.Follower), RequestedAt: follow.CreatedAt})
	}
	return items, buildMeta(page, limit, total), nil
}

func (s *SocialService) ApproveFollowRequest(ctx context.Context, viewerID, requesterID uuid.UUID) error {
	follow, err := s.repo.FindFollow(ctx, requesterID, viewerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrFollowNotFound
	}
	if err != nil {
		return err
	}
	if follow.FollowingID != viewerID {
		return ErrFollowNotFound
	}
	if follow.Status == "accepted" {
		return nil
	}
	return s.repo.UpdateFollowStatus(ctx, requesterID, viewerID, "accepted")
}

func (s *SocialService) DeclineFollowRequest(ctx context.Context, viewerID, requesterID uuid.UUID) error {
	follow, err := s.repo.FindFollow(ctx, requesterID, viewerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if follow.FollowingID != viewerID {
		return ErrFollowNotFound
	}
	return s.repo.DeleteFollow(ctx, requesterID, viewerID)
}

func (s *SocialService) Block(ctx context.Context, viewerID, targetID uuid.UUID) (*dto.BlockStatusResponse, error) {
	if viewerID == targetID {
		return nil, ErrSelfAction
	}
	if err := s.ensureUserExists(ctx, targetID); err != nil {
		return nil, err
	}
	if err := s.repo.CreateBlock(ctx, &model.Block{BlockerID: viewerID, BlockedID: targetID}); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteFollowsBetween(ctx, viewerID, targetID); err != nil {
		return nil, err
	}
	return &dto.BlockStatusResponse{TargetID: targetID.String(), Blocked: true}, nil
}

func (s *SocialService) Unblock(ctx context.Context, viewerID, targetID uuid.UUID) error {
	if viewerID == targetID {
		return ErrSelfAction
	}
	if err := s.ensureUserExists(ctx, targetID); err != nil {
		return err
	}
	return s.repo.DeleteBlock(ctx, viewerID, targetID)
}

func (s *SocialService) Report(ctx context.Context, reporterID uuid.UUID, req dto.ReportRequest) (*dto.ReportResponse, error) {
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return nil, ErrInvalidInput
	}
	targetType := strings.ToLower(strings.TrimSpace(req.TargetType))
	reason := strings.ToLower(strings.TrimSpace(req.Reason))
	if !validReason(reason) {
		return nil, ErrInvalidReason
	}
	if err := s.validateReportTarget(ctx, reporterID, targetType, targetID); err != nil {
		return nil, err
	}
	report := &model.Report{ReporterID: reporterID, TargetType: targetType, TargetID: targetID, Reason: reason, Status: "pending"}
	if err := s.repo.CreateReport(ctx, report); err != nil {
		return nil, err
	}
	return &dto.ReportResponse{ID: report.ID.String(), TargetType: report.TargetType, TargetID: report.TargetID.String(), Reason: report.Reason, Status: report.Status, CreatedAt: report.CreatedAt}, nil
}

func (s *SocialService) GetBlockedUsers(ctx context.Context, viewerID uuid.UUID) ([]dto.SocialUserResponse, error) {
	blocks, err := s.repo.FindBlocksByBlocker(ctx, viewerID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SocialUserResponse, 0, len(blocks))
	for _, b := range blocks {
		out = append(out, buildSocialUser(b.Blocked))
	}
	return out, nil
}

func (s *SocialService) AddCloseFriend(ctx context.Context, userID, targetID uuid.UUID) error {
	if userID == targetID {
		return ErrSelfAction
	}
	if err := s.ensureUserExists(ctx, targetID); err != nil {
		return err
	}
	return s.repo.SetCloseFriend(ctx, userID, targetID, true)
}

func (s *SocialService) RemoveCloseFriend(ctx context.Context, userID, targetID uuid.UUID) error {
	if userID == targetID {
		return ErrSelfAction
	}
	if err := s.ensureUserExists(ctx, targetID); err != nil {
		return err
	}
	return s.repo.SetCloseFriend(ctx, userID, targetID, false)
}

func (s *SocialService) ListCloseFriends(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.SocialUserResponse, *response.Meta, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, nil, err
	}
	page, limit = normalizePagination(page, limit)
	users, total, err := s.repo.ListCloseFriends(ctx, userID, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return buildSocialUsers(users), buildMeta(page, limit, total), nil
}

func (s *SocialService) EnsureProfileVisible(ctx context.Context, viewerID, targetID uuid.UUID) error {
	if viewerID == uuid.Nil || viewerID == targetID {
		return nil
	}
	blocked, err := s.repo.IsBlockedEitherDirection(ctx, viewerID, targetID)
	if err != nil {
		return err
	}
	if blocked {
		return ErrUserBlocked
	}
	return nil
}

func (s *SocialService) ensureUserExists(ctx context.Context, userID uuid.UUID) error {
	exists, err := s.repo.UserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}
	return nil
}

func (s *SocialService) validateReportTarget(ctx context.Context, reporterID uuid.UUID, targetType string, targetID uuid.UUID) error {
	switch targetType {
	case "user":
		if reporterID == targetID {
			return ErrSelfAction
		}
		exists, err := s.repo.UserExists(ctx, targetID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrTargetNotFound
		}
	case "post":
		exists, err := s.repo.PostExists(ctx, targetID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrTargetNotFound
		}
	case "comment":
		exists, err := s.repo.CommentExists(ctx, targetID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrTargetNotFound
		}
	default:
		return ErrInvalidTarget
	}
	return nil
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func buildMeta(page, limit int, total int64) *response.Meta {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return &response.Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}

func buildSocialUsers(users []model.User) []dto.SocialUserResponse {
	items := make([]dto.SocialUserResponse, 0, len(users))
	for _, user := range users {
		items = append(items, buildSocialUser(user))
	}
	return items
}

func buildSocialUser(user model.User) dto.SocialUserResponse {
	return dto.SocialUserResponse{ID: user.ID.String(), Name: user.Name, Username: user.Username, AvatarURL: user.AvatarURL, CreatedAt: user.CreatedAt}
}

func validReason(reason string) bool {
	switch reason {
	case "spam", "inappropriate", "harassment", "fake_account", "other":
		return true
	default:
		return false
	}
}
