package dto

import "time"

type UpdateProfileRequest struct {
	Name         *string `json:"name"`
	Username     *string `json:"username"`
	Bio          *string `json:"bio"`
	AvatarURL    *string `json:"avatar_url"`
	ExternalLink *string `json:"external_link"`
}

type UpdatePrivacyRequest struct {
	IsPrivate *bool `json:"is_private"`
}

type UserProfileResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Username       string     `json:"username"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	PhoneVerified  *bool      `json:"phone_verified,omitempty"`
	EmailVerified  *bool      `json:"email_verified,omitempty"`
	Bio            *string    `json:"bio,omitempty"`
	AvatarURL      *string    `json:"avatar_url,omitempty"`
	ExternalLink   *string    `json:"external_link,omitempty"`
	IsPrivate      bool       `json:"is_private"`
	IsLimited      bool       `json:"is_limited"`
	FollowersCount int64      `json:"followers_count"`
	FollowingCount int64      `json:"following_count"`
	PostsCount     int64      `json:"posts_count"`
	IsFollowing    bool       `json:"is_following"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}
