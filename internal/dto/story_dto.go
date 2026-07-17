package dto

import "time"

type CreateStoryRequest struct {
	MediaURL      *string  `json:"media_url"`
	MediaType     string   `json:"media_type" binding:"required"`
	TextContent   *string  `json:"text_content"`
	MusicTrackURL *string  `json:"music_track_url"`
	TaggedUserIDs []string `json:"tagged_user_ids"`
}

type StoryResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	MediaURL      *string   `json:"media_url,omitempty"`
	MediaType     string    `json:"media_type"`
	TextContent   *string   `json:"text_content,omitempty"`
	MusicTrackURL *string   `json:"music_track_url,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type StoryViewerResponse struct {
	ViewerID  string    `json:"viewer_id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	ViewedAt  time.Time `json:"viewed_at"`
}

type StoryFeedItem struct {
	UserID    string          `json:"user_id"`
	Username  string          `json:"username"`
	AvatarURL *string         `json:"avatar_url,omitempty"`
	Stories   []StoryResponse `json:"stories"`
}
