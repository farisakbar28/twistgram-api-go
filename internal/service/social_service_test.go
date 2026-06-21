package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
)

type fakeSocialRepo struct {
	users          map[uuid.UUID]*model.User
	follows        map[[2]uuid.UUID]*model.Follow
	blocksEither   bool
	blockCreated   bool
	deletedBetween bool
	postExists     bool
	commentExists  bool
	reports        []*model.Report
}

func (f *fakeSocialRepo) FindUserByID(id uuid.UUID) (*model.User, error) {
	if user, ok := f.users[id]; ok {
		copyUser := *user
		return &copyUser, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeSocialRepo) FindFollow(followerID, followingID uuid.UUID) (*model.Follow, error) {
	follow, ok := f.follows[[2]uuid.UUID{followerID, followingID}]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copyFollow := *follow
	return &copyFollow, nil
}
func (f *fakeSocialRepo) UpsertFollow(follow *model.Follow) error {
	if f.follows == nil {
		f.follows = map[[2]uuid.UUID]*model.Follow{}
	}
	copyFollow := *follow
	f.follows[[2]uuid.UUID{follow.FollowerID, follow.FollowingID}] = &copyFollow
	return nil
}
func (f *fakeSocialRepo) DeleteFollow(followerID, followingID uuid.UUID) error {
	delete(f.follows, [2]uuid.UUID{followerID, followingID})
	return nil
}
func (f *fakeSocialRepo) DeleteFollowsBetween(userA, userB uuid.UUID) error {
	f.deletedBetween = true
	return nil
}
func (f *fakeSocialRepo) ListFollowers(userID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeSocialRepo) ListFollowing(userID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeSocialRepo) ListIncomingFollowRequests(userID uuid.UUID, page, limit int) ([]model.Follow, int64, error) {
	return nil, 0, nil
}
func (f *fakeSocialRepo) UpdateFollowStatus(followerID, followingID uuid.UUID, status string) error {
	follow, ok := f.follows[[2]uuid.UUID{followerID, followingID}]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	follow.Status = status
	return nil
}
func (f *fakeSocialRepo) IsBlockedEitherDirection(userA, userB uuid.UUID) (bool, error) {
	return f.blocksEither, nil
}
func (f *fakeSocialRepo) FindBlock(blockerID, blockedID uuid.UUID) (*model.Block, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeSocialRepo) CreateBlock(block *model.Block) error             { f.blockCreated = true; return nil }
func (f *fakeSocialRepo) DeleteBlock(blockerID, blockedID uuid.UUID) error { return nil }
func (f *fakeSocialRepo) CreateReport(report *model.Report) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	f.reports = append(f.reports, report)
	return nil
}
func (f *fakeSocialRepo) UserExists(id uuid.UUID) (bool, error)    { _, ok := f.users[id]; return ok, nil }
func (f *fakeSocialRepo) PostExists(id uuid.UUID) (bool, error)    { return f.postExists, nil }
func (f *fakeSocialRepo) CommentExists(id uuid.UUID) (bool, error) { return f.commentExists, nil }

func TestFollowPrivateUserCreatesPendingRequest(t *testing.T) {
	viewerID := uuid.New()
	targetID := uuid.New()
	repo := &fakeSocialRepo{users: map[uuid.UUID]*model.User{targetID: {ID: targetID, IsPrivate: true}}}
	svc := NewSocialService(repo)

	result, err := svc.Follow(viewerID, targetID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Status != "pending" {
		t.Fatalf("expected pending status, got %s", result.Status)
	}
	if repo.follows[[2]uuid.UUID{viewerID, targetID}].Status != "pending" {
		t.Fatal("expected pending follow stored")
	}
}

func TestFollowBlockedUserFails(t *testing.T) {
	viewerID := uuid.New()
	targetID := uuid.New()
	repo := &fakeSocialRepo{users: map[uuid.UUID]*model.User{targetID: {ID: targetID}}, blocksEither: true}
	svc := NewSocialService(repo)

	_, err := svc.Follow(viewerID, targetID)
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("expected ErrBlocked, got %v", err)
	}
}

func TestBlockDeletesFollowsBothDirections(t *testing.T) {
	viewerID := uuid.New()
	targetID := uuid.New()
	repo := &fakeSocialRepo{users: map[uuid.UUID]*model.User{targetID: {ID: targetID}}}
	svc := NewSocialService(repo)

	_, err := svc.Block(viewerID, targetID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !repo.blockCreated || !repo.deletedBetween {
		t.Fatal("expected block created and follows deleted")
	}
}

func TestReportValidatesUserSelfReport(t *testing.T) {
	userID := uuid.New()
	repo := &fakeSocialRepo{users: map[uuid.UUID]*model.User{userID: {ID: userID}}}
	svc := NewSocialService(repo)

	_, err := svc.Report(userID, dto.ReportRequest{TargetType: "user", TargetID: userID.String(), Reason: "spam"})
	if !errors.Is(err, ErrSelfAction) {
		t.Fatalf("expected ErrSelfAction, got %v", err)
	}
}

func TestReportPostTargetMustExist(t *testing.T) {
	reporterID := uuid.New()
	postID := uuid.New()
	repo := &fakeSocialRepo{users: map[uuid.UUID]*model.User{}}
	svc := NewSocialService(repo)

	_, err := svc.Report(reporterID, dto.ReportRequest{TargetType: "post", TargetID: postID.String(), Reason: "spam"})
	if !errors.Is(err, ErrTargetNotFound) {
		t.Fatalf("expected ErrTargetNotFound, got %v", err)
	}
}
