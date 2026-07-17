package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversation_id"`
	SenderID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"sender_id"`
	MessageType      string     `gorm:"type:varchar(20);not null;default:'TEXT'" json:"message_type"`
	Content          *string    `gorm:"type:text;null" json:"content,omitempty"`
	MediaURL         *string    `gorm:"type:varchar(500);null" json:"media_url,omitempty"`
	ReplyToStoryID   *uuid.UUID `gorm:"type:uuid;null;index" json:"reply_to_story_id,omitempty"`
	ReplyToMessageID *uuid.UUID `gorm:"type:uuid;null;index" json:"reply_to_message_id,omitempty"`
	IsRead           bool       `gorm:"default:false" json:"is_read"`
	CreatedAt        time.Time  `json:"created_at"`

	Conversation   Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	Sender         User         `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReplyToStory   *Story       `gorm:"foreignKey:ReplyToStoryID" json:"reply_to_story,omitempty"`
	ReplyToMessage *Message     `gorm:"foreignKey:ReplyToMessageID" json:"reply_to_message,omitempty"`
}

func (Message) TableName() string {
	return "messages"
}
