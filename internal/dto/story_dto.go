package dto

import "time"

type CreateStoryRequest struct {
	MediaType          string   `json:"media_type" binding:"required"`
	MediaURL           *string  `json:"media_url"`
	TextContent        *string  `json:"text_content"`
	MusicTrackURL      *string  `json:"music_track_url"`
	IsCloseFriendsOnly bool     `json:"is_close_friends_only"`
	TaggedUserIDs      []string `json:"tagged_user_ids"`
}

type StoryResponse struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	MediaURL           *string   `json:"media_url,omitempty"`
	MediaType          string    `json:"media_type"`
	TextContent        *string   `json:"text_content,omitempty"`
	MusicTrackURL      *string   `json:"music_track_url,omitempty"`
	IsCloseFriendsOnly bool      `json:"is_close_friends_only"`
	ExpiresAt          time.Time `json:"expires_at"`
	CreatedAt          time.Time `json:"created_at"`
}

type StoryFeedItem struct {
	UserID    string          `json:"user_id"`
	Username  string          `json:"username"`
	AvatarURL *string         `json:"avatar_url,omitempty"`
	Stories   []StoryResponse `json:"stories"`
}

type StoryViewerResponse struct {
	ViewerID  string    `json:"viewer_id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	ViewedAt  time.Time `json:"viewed_at"`
}
