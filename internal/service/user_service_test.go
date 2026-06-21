package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
)

type fakeUserRepo struct {
	user           *model.User
	usernameExists bool
	updated        *model.User
}

func (f *fakeUserRepo) FindByID(id uuid.UUID) (*model.User, error) {
	if f.user == nil || f.user.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	copyUser := *f.user
	return &copyUser, nil
}

func (f *fakeUserRepo) FindByUsername(username string) (*model.User, error) {
	if f.user == nil || f.user.Username != username {
		return nil, gorm.ErrRecordNotFound
	}
	copyUser := *f.user
	return &copyUser, nil
}

func (f *fakeUserRepo) UsernameExists(username string, excludeID uuid.UUID) (bool, error) {
	return f.usernameExists, nil
}

func (f *fakeUserRepo) Update(user *model.User) error {
	copyUser := *user
	f.updated = &copyUser
	f.user = &copyUser
	return nil
}

func (f *fakeUserRepo) CountFollowers(userID uuid.UUID) (int64, error) { return 1, nil }
func (f *fakeUserRepo) CountFollowing(userID uuid.UUID) (int64, error) { return 2, nil }
func (f *fakeUserRepo) CountPosts(userID uuid.UUID) (int64, error)     { return 3, nil }
func (f *fakeUserRepo) IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error) {
	return false, nil
}

func TestUpdateProfileUsernameChangeLimited(t *testing.T) {
	userID := uuid.New()
	lastUsernameAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeUserRepo{user: &model.User{
		ID:             userID,
		Name:           "User",
		Username:       "old_name",
		Email:          "user@example.com",
		LastUsernameAt: &lastUsernameAt,
	}}
	service := NewUserServiceWithClock(repo, func() time.Time {
		return lastUsernameAt.AddDate(0, 0, 10)
	})
	newUsername := "new_name"

	_, err := service.UpdateProfile(userID, dto.UpdateProfileRequest{Username: &newUsername})
	if !errors.Is(err, ErrUsernameChangeLimited) {
		t.Fatalf("expected ErrUsernameChangeLimited, got %v", err)
	}
	if repo.updated != nil {
		t.Fatal("expected user not to be updated")
	}
}

func TestUpdateProfileUsernameValidation(t *testing.T) {
	badUsername := "Bad-Username"
	err := ValidateUpdateProfileRequest(dto.UpdateProfileRequest{Username: &badUsername})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateProfileUsernameTaken(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{
		user: &model.User{
			ID:       userID,
			Name:     "User",
			Username: "old_name",
			Email:    "user@example.com",
		},
		usernameExists: true,
	}
	service := NewUserServiceWithClock(repo, func() time.Time {
		return time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	})
	newUsername := "new_name"

	_, err := service.UpdateProfile(userID, dto.UpdateProfileRequest{Username: &newUsername})
	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestGetPrivateProfileLimitedForNonFollower(t *testing.T) {
	userID := uuid.New()
	viewerID := uuid.New()
	bio := "secret bio"
	repo := &fakeUserRepo{user: &model.User{
		ID:        userID,
		Name:      "Private User",
		Username:  "private_user",
		Email:     "private@example.com",
		Bio:       &bio,
		IsPrivate: true,
	}}
	service := NewUserService(repo)

	profile, err := service.GetProfileByUsername("private_user", viewerID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !profile.IsLimited {
		t.Fatal("expected limited profile")
	}
	if profile.Email != nil || profile.Phone != nil || profile.Bio != nil {
		t.Fatalf("expected no contact or bio in limited profile, got %+v", profile)
	}
	if profile.PostsCount != 0 {
		t.Fatalf("expected posts count hidden, got %d", profile.PostsCount)
	}
}
