package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
)

type fakeStoryRepo struct {
	stories      map[uuid.UUID]*model.Story
	blocked      bool
	isFollower   bool
	viewers      []model.StoryView
	notifCreated bool
}

func (f *fakeStoryRepo) CreateStoryWithTags(ctx context.Context, story *model.Story, tags []model.StoryTag) error {
	if story.ID == uuid.Nil {
		story.ID = uuid.New()
	}
	if f.stories == nil {
		f.stories = make(map[uuid.UUID]*model.Story)
	}
	f.stories[story.ID] = story
	return nil
}
func (f *fakeStoryRepo) GetStoryByID(ctx context.Context, id uuid.UUID) (*model.Story, error) {
	s, ok := f.stories[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}
func (f *fakeStoryRepo) ListActiveFeedStories(ctx context.Context, userID uuid.UUID) ([]model.Story, error) {
	var active []model.Story
	for _, s := range f.stories {
		if s.ExpiresAt.After(time.Now()) {
			active = append(active, *s)
		}
	}
	return active, nil
}
func (f *fakeStoryRepo) FindActiveStoryByUserID(ctx context.Context, userID uuid.UUID) (*model.Story, error) {
	return nil, nil
}
func (f *fakeStoryRepo) DeleteExpiredStories(ctx context.Context) error {
	return nil
}
func (f *fakeStoryRepo) RecordView(ctx context.Context, view *model.StoryView) error {
	f.viewers = append(f.viewers, *view)
	return nil
}
func (f *fakeStoryRepo) ListViewers(ctx context.Context, storyID uuid.UUID) ([]model.StoryView, error) {
	return f.viewers, nil
}
func (f *fakeStoryRepo) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	return f.blocked, nil
}
func (f *fakeStoryRepo) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return f.isFollower, nil
}
func (f *fakeStoryRepo) DeleteStory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	delete(f.stories, id)
	return nil
}
func (f *fakeStoryRepo) GetStoryOwner(ctx context.Context, storyID uuid.UUID) (uuid.UUID, error) {
	if s, ok := f.stories[storyID]; ok {
		return s.UserID, nil
	}
	return uuid.Nil, errors.New("not found")
}
func (f *fakeStoryRepo) CreateNotification(ctx context.Context, notification *model.Notification) error {
	f.notifCreated = true
	return nil
}

func TestGetStoryByIDBlockedForbidden(t *testing.T) {
	viewerID := uuid.New()
	ownerID := uuid.New()
	storyID := uuid.New()
	repo := &fakeStoryRepo{
		stories: map[uuid.UUID]*model.Story{storyID: {ID: storyID, UserID: ownerID, ExpiresAt: time.Now().Add(time.Hour)}},
		blocked: true,
	}
	svc := NewStoryService(repo)

	_, err := svc.GetByID(context.Background(), viewerID, storyID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for blocked story access, got %v", err)
	}
}

func TestRecordViewOnlyOncePerViewer(t *testing.T) {
	viewerID := uuid.New()
	ownerID := uuid.New()
	storyID := uuid.New()
	repo := &fakeStoryRepo{
		stories: map[uuid.UUID]*model.Story{storyID: {ID: storyID, UserID: ownerID, ExpiresAt: time.Now().Add(time.Hour)}},
	}
	svc := NewStoryService(repo)

	err := svc.RecordView(context.Background(), viewerID, storyID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(repo.viewers) != 1 {
		t.Fatalf("expected 1 view recorded, got %d", len(repo.viewers))
	}
}

func TestCreateStoryWithTags(t *testing.T) {
	userID := uuid.New()
	taggedID := uuid.New()
	repo := &fakeStoryRepo{}
	svc := NewStoryService(repo)

	mediaURL := "https://example.com/story.jpg"
	req := dto.CreateStoryRequest{
		MediaType:     "image",
		MediaURL:      &mediaURL,
		TaggedUserIDs: []string{taggedID.String()},
	}

	res, err := svc.Create(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.ID == "" {
		t.Fatal("expected non-empty story ID")
	}
	if !repo.notifCreated {
		t.Fatal("expected notification sent to tagged user")
	}
}
