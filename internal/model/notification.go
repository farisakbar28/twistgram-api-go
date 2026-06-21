package model

import (
	"time"

	"github.com/google/uuid"
)

// Notification menyimpan notifikasi in-app untuk pengguna.
// type: like, comment, follow, follow_request, mention, story_reply
type Notification struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipientID uuid.UUID  `gorm:"type:uuid;not null;index" json:"recipient_id"`
	ActorID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"actor_id"`
	Type        string     `gorm:"type:varchar(30);not null" json:"type"`
	ReferenceID *uuid.UUID `gorm:"type:uuid;null" json:"reference_id,omitempty"`
	IsRead      bool       `gorm:"default:false" json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`

	// Relations
	Recipient User `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	Actor     User `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
