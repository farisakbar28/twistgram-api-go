package service

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/pkg/auth"
)

var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrUserNotFound          = errors.New("user not found")
	ErrUsernameTaken         = errors.New("username already taken")
	ErrUsernameChangeLimited = errors.New("username can only be changed once per month")
	ErrPasswordRequired      = errors.New("password is required to delete account")
	ErrWrongPassword         = errors.New("incorrect password")
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

type UserService struct {
	repo repository.UserRepository
	now  func() time.Time
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo, now: time.Now}
}

func NewUserServiceWithClock(repo repository.UserRepository, now func() time.Time) *UserService {
	return &UserService{repo: repo, now: now}
}

func (s *UserService) GetMe(userID uuid.UUID) (*dto.UserProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	counts, err := s.counts(user.ID)
	if err != nil {
		return nil, err
	}

	profile := buildProfile(user, counts, false, true, false)
	return &profile, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserProfileResponse, error) {
	if err := ValidateUpdateProfileRequest(req); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = strings.TrimSpace(*req.Name)
	}
	if req.Bio != nil {
		user.Bio = cleanOptionalString(req.Bio)
	}
	if req.AvatarURL != nil {
		user.AvatarURL = cleanOptionalString(req.AvatarURL)
	}
	if req.ExternalLink != nil {
		user.ExternalLink = cleanOptionalString(req.ExternalLink)
	}
	if req.Username != nil {
		newUsername := strings.TrimSpace(strings.ToLower(*req.Username))
		if newUsername != user.Username {
			if user.LastUsernameAt != nil && s.now().Before(user.LastUsernameAt.AddDate(0, 1, 0)) {
				return nil, ErrUsernameChangeLimited
			}
			exists, err := s.repo.UsernameExists(newUsername, user.ID)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrUsernameTaken
			}
			now := s.now()
			user.Username = newUsername
			user.LastUsernameAt = &now
		}
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	counts, err := s.counts(user.ID)
	if err != nil {
		return nil, err
	}
	profile := buildProfile(user, counts, false, true, false)
	return &profile, nil
}

func (s *UserService) GetInterests(userID uuid.UUID) (*dto.UserInterestsResponse, error) {
	if userID == uuid.Nil { return nil, ErrInvalidInput }
	if err := s.ensureUserExists(userID); err != nil { return nil, err }
	interests, err := s.repo.GetInterests(userID)
	if err != nil { return nil, err }
	return &dto.UserInterestsResponse{Interests: interests}, nil
}

func (s *UserService) SetInterests(userID uuid.UUID, req dto.UserInterestsRequest) (*dto.UserInterestsResponse, error) {
	if userID == uuid.Nil { return nil, ErrInvalidInput }
	if err := s.ensureUserExists(userID); err != nil { return nil, err }
	clean := make([]string, 0, len(req.Interests))
	seen := make(map[string]bool)
	for _, cat := range req.Interests {
		c := strings.TrimSpace(strings.ToLower(cat))
		if c != "" && !seen[c] {
			clean = append(clean, c)
			seen[c] = true
		}
	}
	if err := s.repo.SetInterests(userID, clean); err != nil { return nil, err }
	return &dto.UserInterestsResponse{Interests: clean}, nil
}

func (s *UserService) ensureUserExists(userID uuid.UUID) error {
	_, err := s.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) { return ErrUserNotFound }
	return err
}

func (s *UserService) DeleteAccount(userID uuid.UUID, password string) error {
	if userID == uuid.Nil || password == "" {
		return ErrInvalidInput
	}
	user, err := s.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	if user.PasswordHash == nil || !auth.CheckPasswordHash(password, *user.PasswordHash) {
		return ErrWrongPassword
	}
	return s.repo.DeleteUser(userID)
}

func (s *UserService) UpdatePrivacy(userID uuid.UUID, req dto.UpdatePrivacyRequest) (*dto.UserProfileResponse, error) {
	if req.IsPrivate == nil {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.IsPrivate = *req.IsPrivate
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	counts, err := s.counts(user.ID)
	if err != nil {
		return nil, err
	}
	profile := buildProfile(user, counts, false, true, false)
	return &profile, nil
}

func (s *UserService) GetProfileByUsername(username string, viewerID uuid.UUID) (*dto.UserProfileResponse, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	if !usernamePattern.MatchString(username) {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.FindByUsername(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	isSelf := viewerID != uuid.Nil && viewerID == user.ID
	isFollowing := false
	if viewerID != uuid.Nil && !isSelf {
		blocked, err := s.repo.IsBlockedEitherDirection(viewerID, user.ID)
		if err != nil {
			return nil, err
		}
		if blocked {
			return nil, ErrUserBlocked
		}
		isFollowing, err = s.repo.IsAcceptedFollower(viewerID, user.ID)
		if err != nil {
			return nil, err
		}
	}
	limited := user.IsPrivate && !isSelf && !isFollowing

	counts, err := s.counts(user.ID)
	if err != nil {
		return nil, err
	}
	profile := buildProfile(user, counts, limited, false, isFollowing)
	return &profile, nil
}

func ValidateUpdateProfileRequest(req dto.UpdateProfileRequest) error {
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 1 || len(name) > 100 {
			return ErrInvalidInput
		}
	}
	if req.Username != nil {
		username := strings.TrimSpace(strings.ToLower(*req.Username))
		if !usernamePattern.MatchString(username) {
			return ErrInvalidInput
		}
	}
	if req.Bio != nil && len(strings.TrimSpace(*req.Bio)) > 500 {
		return ErrInvalidInput
	}
	if req.AvatarURL != nil {
		avatar := strings.TrimSpace(*req.AvatarURL)
		if avatar != "" && (len(avatar) > 500 || !isValidHTTPURL(avatar)) {
			return ErrInvalidInput
		}
	}
	if req.ExternalLink != nil {
		link := strings.TrimSpace(*req.ExternalLink)
		if link != "" && (len(link) > 500 || !isValidHTTPURL(link)) {
			return ErrInvalidInput
		}
	}
	return nil
}

type profileCounts struct {
	followers int64
	following int64
	posts     int64
}

func (s *UserService) counts(userID uuid.UUID) (profileCounts, error) {
	followers, err := s.repo.CountFollowers(userID)
	if err != nil {
		return profileCounts{}, err
	}
	following, err := s.repo.CountFollowing(userID)
	if err != nil {
		return profileCounts{}, err
	}
	posts, err := s.repo.CountPosts(userID)
	if err != nil {
		return profileCounts{}, err
	}
	return profileCounts{followers: followers, following: following, posts: posts}, nil
}

func buildProfile(user *model.User, counts profileCounts, limited, includeContact, isFollowing bool) dto.UserProfileResponse {
	profile := dto.UserProfileResponse{
		ID:             user.ID.String(),
		Name:           user.Name,
		Username:       user.Username,
		Bio:            user.Bio,
		AvatarURL:      user.AvatarURL,
		ExternalLink:   user.ExternalLink,
		IsPrivate:      user.IsPrivate,
		IsLimited:      limited,
		FollowersCount: counts.followers,
		FollowingCount: counts.following,
		PostsCount:     counts.posts,
		IsFollowing:    isFollowing,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      &user.UpdatedAt,
	}
	if includeContact {
		profile.Email = &user.Email
		profile.Phone = user.Phone
		profile.PhoneVerified = &user.PhoneVerified
		profile.EmailVerified = &user.EmailVerified
	}
	if limited {
		profile.Bio = nil
		profile.ExternalLink = nil
		profile.PostsCount = 0
		profile.UpdatedAt = nil
	}
	return profile
}

func cleanOptionalString(value *string) *string {
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isValidHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
