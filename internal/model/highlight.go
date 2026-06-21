package model

import (
	"time"

	"github.com/google/uuid"
)

// Highlight adalah koleksi story permanen di profil. [ADV]
type Highlight struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Title     string    `gorm:"type:varchar(100);not null" json:"title"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User   User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Stories []HighlightStory `gorm:"foreignKey:HighlightID" json:"stories,omitempty"`
}

func (Highlight) TableName() string {
	return "highlights"
}
