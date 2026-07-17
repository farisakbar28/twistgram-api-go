package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
)

type fakePostRepo struct {
	posts        map[uuid.UUID]*model.Post
	blocked      bool
	isPrivate    bool
	isFollower   bool
	postCreated  bool
	notifCreated bool
}

func (f *fakePostRepo) CreatePostWithMediaAndTags(ctx context.Context, post *model.Post, media []model.PostMedia, tags []model.PostTag, hashtags []model.Hashtag) error {
	if post.ID == uuid.Nil {
		post.ID = uuid.New()
	}
	if f.posts == nil {
		f.posts = make(map[uuid.UUID]*model.Post)
	}
	f.posts[post.ID] = post
	f.postCreated = true
	return nil
}
func (f *fakePostRepo) UpdateCaption(ctx context.Context, id uuid.UUID, userID uuid.UUID, caption *string) error {
	p, ok := f.posts[id]
	if !ok || p.UserID != userID {
		return errors.New("not found")
	}
	p.Caption = caption
	return nil
}
func (f *fakePostRepo) GetPostByID(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	p, ok := f.posts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}
func (f *fakePostRepo) GetPostWithMedia(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	return f.GetPostByID(ctx, id)
}
func (f *fakePostRepo) UpdatePost(ctx context.Context, post *model.Post) error {
	f.posts[post.ID] = post
	return nil
}
func (f *fakePostRepo) ListFeed(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Post, int64, error) {
	return nil, 0, nil
}
func (f *fakePostRepo) ListUserPosts(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Post, int64, error) {
	var list []model.Post
	for _, p := range f.posts {
		if p.UserID == userID {
			list = append(list, *p)
		}
	}
	return list, int64(len(list)), nil
}
func (f *fakePostRepo) ListGlobalFeed(ctx context.Context, viewerID uuid.UUID, limit int) ([]model.Post, int64, error) {
	return nil, 0, nil
}
func (f *fakePostRepo) DeletePost(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	delete(f.posts, id)
	return nil
}
func (f *fakePostRepo) SetArchived(ctx context.Context, id uuid.UUID, userID uuid.UUID, archived bool) error {
	if p, ok := f.posts[id]; ok && p.UserID == userID {
		p.IsArchived = archived
		return nil
	}
	return errors.New("not found")
}
func (f *fakePostRepo) PostExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := f.posts[id]
	return ok, nil
}
func (f *fakePostRepo) OwnsPost(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	if p, ok := f.posts[id]; ok {
		return p.UserID == userID, nil
	}
	return false, nil
}
func (f *fakePostRepo) UserHashtagUpsert(ctx context.Context, tags []string) ([]model.Hashtag, error) {
	return nil, nil
}
func (f *fakePostRepo) CreateNotification(ctx context.Context, notification *model.Notification) error {
	f.notifCreated = true
	return nil
}
func (f *fakePostRepo) DeletePostTag(ctx context.Context, postID, taggedUserID uuid.UUID) error {
	return nil
}
func (f *fakePostRepo) IsUserPrivate(ctx context.Context, userID uuid.UUID) (bool, error) {
	return f.isPrivate, nil
}
func (f *fakePostRepo) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	return f.blocked, nil
}
func (f *fakePostRepo) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return f.isFollower, nil
}

func TestUserPostsBlockedForbidden(t *testing.T) {
	viewerID := uuid.New()
	targetID := uuid.New()
	repo := &fakePostRepo{blocked: true}
	svc := NewPostService(repo)

	_, _, err := svc.UserPosts(context.Background(), viewerID, targetID, 1, 20)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for blocked user posts, got %v", err)
	}
}

func TestUserPostsPrivateNonFollowerForbidden(t *testing.T) {
	viewerID := uuid.New()
	targetID := uuid.New()
	repo := &fakePostRepo{isPrivate: true, isFollower: false}
	svc := NewPostService(repo)

	_, _, err := svc.UserPosts(context.Background(), viewerID, targetID, 1, 20)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for private account posts when non-follower, got %v", err)
	}
}

func TestCreatePostWithMedia(t *testing.T) {
	userID := uuid.New()
	repo := &fakePostRepo{}
	svc := NewPostService(repo)

	cap := "Hello world #twistgram"
	req := dto.CreatePostRequest{
		Caption: &cap,
		Media: []dto.PostMediaRequest{
			{MediaURL: "https://example.com/a.jpg", MediaType: "image"},
		},
	}

	res, err := svc.CreatePost(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.User.UserID != userID.String() {
		t.Fatalf("expected userID %s, got %s", userID.String(), res.User.UserID)
	}
	if !repo.postCreated {
		t.Fatal("expected post created in repository")
	}
}

