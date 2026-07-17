package dto

import "time"

type PostMediaRequest struct {
	MediaURL      string  `json:"media_url" binding:"required"`
	MediaType     string  `json:"media_type" binding:"required"`
	OrderIndex    int     `json:"order_index"`
	MusicTrackURL *string `json:"music_track_url"`
}

type CreatePostRequest struct {
	Caption          *string            `json:"caption"`
	Location         *string            `json:"location"`
	CommentsDisabled *bool              `json:"comments_disabled"`
	Media            []PostMediaRequest `json:"media" binding:"required,min=1,max=10"`
	TaggedUserIDs    []string           `json:"tagged_user_ids"`
	Hashtags         []string           `json:"hashtags"`
}

type UpdatePostRequest struct {
	Caption *string `json:"caption"`
}

type PostMediaResponse struct {
	MediaURL      string  `json:"media_url"`
	MediaType     string  `json:"media_type"`
	OrderIndex    int     `json:"order_index"`
	MusicTrackURL *string `json:"music_track_url,omitempty"`
}

type PostResponse struct {
	ID               string              `json:"id"`
	UserID           string              `json:"user_id"`
	Caption          *string             `json:"caption,omitempty"`
	Location         *string             `json:"location,omitempty"`
	IsArchived       bool                `json:"is_archived"`
	CommentsDisabled bool                `json:"comments_disabled"`
	Media            []PostMediaResponse `json:"media,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
}

type FeedResponse struct {
	Posts []PostResponse `json:"posts"`
}
