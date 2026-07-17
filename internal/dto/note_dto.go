package dto

import "time"

type CreateNoteRequest struct {
	Content      string  `json:"content" binding:"required,max=60"`
	AudioTrackID *string `json:"audio_track_id"`
}

type NoteResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	Content      string    `json:"content"`
	AudioTrackID *string   `json:"audio_track_id,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
