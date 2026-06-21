package model

import (
	"time"

	"github.com/google/uuid"
)

// Message menyimpan pesan dalam percakapan.
// reply_to_story_id: penghubung Story Reply yang masuk sebagai DM.
type Message struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversation_id"`
	SenderID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"sender_id"`
	Content         *string    `gorm:"type:text;null" json:"content,omitempty"`
	MediaURL        *string    `gorm:"type:varchar(500);null" json:"media_url,omitempty"`
	ReplyToStoryID  *uuid.UUID `gorm:"type:uuid;null;index" json:"reply_to_story_id,omitempty"`
	IsRead          bool       `gorm:"default:false" json:"is_read"` // [ADV]
	CreatedAt       time.Time  `json:"created_at"`

	// Relations
	Conversation   Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	Sender         User         `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReplyToStory   *Story       `gorm:"foreignKey:ReplyToStoryID" json:"reply_to_story,omitempty"`
}

func (Message) TableName() string {
	return "messages"
}
