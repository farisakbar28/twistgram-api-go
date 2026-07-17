package dto

import "time"

type StartConversationRequest struct {
	TargetID string `json:"target_id" binding:"required"`
}

type CreateGroupRequest struct {
	GroupTitle     string   `json:"group_title" binding:"required"`
	ParticipantIDs []string `json:"participant_ids" binding:"required,min=2,max=100"`
}

type MessageRequest struct {
	Content          *string `json:"content"`
	MediaURL         *string `json:"media_url"`
	StoryID          *string `json:"story_id"`
	ReplyToMessageID *string `json:"reply_to_message_id"`
}

type ConversationResponse struct {
	ID         string          `json:"id"`
	IsGroup    bool            `json:"is_group"`
	GroupTitle *string         `json:"group_title,omitempty"`
	OtherUser  *SearchUserItem `json:"other_user,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type MessageResponse struct {
	ID               string    `json:"id"`
	ConversationID   string    `json:"conversation_id"`
	SenderID         string    `json:"sender_id"`
	MessageType      string    `json:"message_type"`
	Content          *string   `json:"content,omitempty"`
	MediaURL         *string   `json:"media_url,omitempty"`
	ReplyToStoryID   *string   `json:"reply_to_story_id,omitempty"`
	ReplyToMessageID *string   `json:"reply_to_message_id,omitempty"`
	IsRead           bool      `json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}
