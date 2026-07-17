package dto

import "time"

type StartConversationRequest struct {
	TargetID string `json:"target_id" binding:"required"`
}

type ConversationResponse struct {
	ID        string          `json:"id"`
	IsGroup   bool            `json:"is_group"`
	OtherUser *SearchUserItem `json:"other_user,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type MessageRequest struct {
	Content  *string `json:"content"`
	MediaURL *string `json:"media_url"`
	StoryID  *string `json:"story_id"`
}

type MessageResponse struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Content        *string   `json:"content,omitempty"`
	MediaURL       *string   `json:"media_url,omitempty"`
	ReplyToStoryID *string   `json:"reply_to_story_id,omitempty"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

type NotificationResponse struct {
	ID          string    `json:"id"`
	RecipientID string    `json:"recipient_id"`
	ActorID     string    `json:"actor_id"`
	Type        string    `json:"type"`
	ReferenceID *string   `json:"reference_id,omitempty"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}
