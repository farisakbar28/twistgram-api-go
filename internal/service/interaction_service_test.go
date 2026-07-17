package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

type fakeInteractionRepo struct {
	posts        map[uuid.UUID]*model.Post
	comments     map[uuid.UUID]*model.Comment
	likes        map[string]bool
	saved        map[string]bool
	blocked      bool
	isPrivate    bool
	isFollower   bool
	notifCreated bool
}

func (f *fakeInteractionRepo) PostExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := f.posts[id]
	return ok, nil
}
func (f *fakeInteractionRepo) CommentExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := f.comments[id]
	return ok, nil
}
func (f *fakeInteractionRepo) FindCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	c, ok := f.comments[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}
func (f *fakeInteractionRepo) CreateComment(ctx context.Context, comment *model.Comment) error {
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	if f.comments == nil {
		f.comments = make(map[uuid.UUID]*model.Comment)
	}
	f.comments[comment.ID] = comment
	return nil
}
func (f *fakeInteractionRepo) DeleteComment(ctx context.Context, id, userID uuid.UUID) error {
	if c, ok := f.comments[id]; ok && c.UserID == userID {
		delete(f.comments, id)
		return nil
	}
	return errors.New("not found")
}
func (f *fakeInteractionRepo) DeleteCommentAsPostOwner(ctx context.Context, commentID, postID uuid.UUID) error {
	delete(f.comments, commentID)
	return nil
}
func (f *fakeInteractionRepo) UpsertLike(ctx context.Context, like *model.Like) error {
	if f.likes == nil {
		f.likes = make(map[string]bool)
	}
	f.likes[like.UserID.String()+like.LikeableType+like.LikeableID.String()] = true
	return nil
}
func (f *fakeInteractionRepo) DeleteLike(ctx context.Context, userID uuid.UUID, likeableType string, likeableID uuid.UUID) error {
	delete(f.likes, userID.String()+likeableType+likeableID.String())
	return nil
}
func (f *fakeInteractionRepo) SavedExists(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	return f.saved[userID.String()+postID.String()], nil
}
func (f *fakeInteractionRepo) UpsertSavedPost(ctx context.Context, saved *model.SavedPost) error {
	if f.saved == nil {
		f.saved = make(map[string]bool)
	}
	f.saved[saved.UserID.String()+saved.PostID.String()] = true
	return nil
}
func (f *fakeInteractionRepo) DeleteSavedPost(ctx context.Context, userID, postID uuid.UUID) error {
	delete(f.saved, userID.String()+postID.String())
	return nil
}
func (f *fakeInteractionRepo) ListPostComments(ctx context.Context, postID uuid.UUID, page, limit int) ([]model.Comment, int64, error) {
	var list []model.Comment
	for _, c := range f.comments {
		if c.PostID == postID {
			list = append(list, *c)
		}
	}
	return list, int64(len(list)), nil
}
func (f *fakeInteractionRepo) ListSavedPosts(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.SavedPost, int64, error) {
	return nil, 0, nil
}
func (f *fakeInteractionRepo) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	return f.blocked, nil
}
func (f *fakeInteractionRepo) GetPostOwner(ctx context.Context, postID uuid.UUID) (*repository.PostOwnerInfo, error) {
	if p, ok := f.posts[postID]; ok {
		return &repository.PostOwnerInfo{UserID: p.UserID, IsPrivate: f.isPrivate}, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeInteractionRepo) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return f.isFollower, nil
}
func (f *fakeInteractionRepo) GetCommentPostID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error) {
	if c, ok := f.comments[commentID]; ok {
		return c.PostID, nil
	}
	return uuid.Nil, errors.New("not found")
}
func (f *fakeInteractionRepo) CreateNotification(ctx context.Context, notification *model.Notification) error {
	f.notifCreated = true
	return nil
}

func TestLikePostNotifiesOwner(t *testing.T) {
	userID := uuid.New()
	ownerID := uuid.New()
	postID := uuid.New()
	repo := &fakeInteractionRepo{
		posts: map[uuid.UUID]*model.Post{postID: {ID: postID, UserID: ownerID}},
	}
	svc := NewInteractionService(repo)

	res, err := svc.LikePost(context.Background(), userID, postID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !res.Liked {
		t.Fatal("expected post to be liked")
	}
	if !repo.notifCreated {
		t.Fatal("expected notification created for post owner")
	}
}

func TestCommentOnBlockedPostFails(t *testing.T) {
	userID := uuid.New()
	ownerID := uuid.New()
	postID := uuid.New()
	repo := &fakeInteractionRepo{
		posts:   map[uuid.UUID]*model.Post{postID: {ID: postID, UserID: ownerID}},
		blocked: true,
	}
	svc := NewInteractionService(repo)

	_, err := svc.CreateComment(context.Background(), userID, postID, dto.CreateCommentRequest{Content: "Hello!"})
	if !errors.Is(err, ErrUserBlocked) {
		t.Fatalf("expected ErrUserBlocked, got %v", err)
	}
}
