package dto

import "time"

type CreateHighlightRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateHighlightRequest struct {
	Title string `json:"title" binding:"required"`
}

type AddStoryToHighlightRequest struct {
	StoryID string `json:"story_id" binding:"required"`
}

type HighlightResponse struct {
	ID        string                    `json:"id"`
	Title     string                    `json:"title"`
	Stories   []HighlightStoryResponse  `json:"stories,omitempty"`
	CreatedAt time.Time                 `json:"created_at"`
}

type HighlightStoryResponse struct {
	ID        string    `json:"id"`
	StoryID   string    `json:"story_id"`
	MediaURL  *string   `json:"media_url,omitempty"`
	MediaType string    `json:"media_type"`
	CreatedAt time.Time `json:"created_at"`
}
