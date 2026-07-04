package dto

import "time"

type CreateCommentRequest struct {
	Content          string  `json:"content" binding:"required"`
	ParentCommentID  *string `json:"parent_comment_id"`
}

type CommentResponse struct {
	ID              string     `json:"id"`
	PostID          string     `json:"post_id"`
	UserID          string     `json:"user_id"`
	ParentCommentID *string    `json:"parent_comment_id,omitempty"`
	Content         string     `json:"content"`
	IsPinned        bool       `json:"is_pinned"`
	CreatedAt       time.Time  `json:"created_at"`
}

type LikeStatusResponse struct {
	TargetID string `json:"target_id"`
	Liked    bool   `json:"liked"`
}

type SaveStatusResponse struct {
	PostID string `json:"post_id"`
	Saved  bool   `json:"saved"`
}

type SavedPostResponse struct {
	ID        string       `json:"id"`
	PostID    string       `json:"post_id"`
	Collection string      `json:"collection_name"`
	CreatedAt time.Time    `json:"created_at"`
	// Minimal post detail
	Caption   *string      `json:"caption,omitempty"`
}

type ShareResponse struct {
	Link string `json:"link"`
}
