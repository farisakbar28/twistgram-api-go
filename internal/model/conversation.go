package model

import (
	"time"

	"github.com/google/uuid"
)

// Conversation adalah percakapan antar pengguna. [ADV] is_group untuk group chat.
type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IsGroup   bool      `gorm:"default:false" json:"is_group"` // [ADV]
	CreatedAt time.Time `json:"created_at"`

	// Relations
	Participants []ConversationParticipant `gorm:"foreignKey:ConversationID" json:"participants,omitempty"`
	Messages     []Message                 `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// ConversationParticipant menyimpan partisipan dalam suatu percakapan.
type ConversationParticipant struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cnv_part" json:"conversation_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cnv_part" json:"user_id"`

	// Relations
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ConversationParticipant) TableName() string {
	return "conversation_participants"
}
