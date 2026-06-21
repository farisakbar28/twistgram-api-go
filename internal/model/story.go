package model

import (
	"time"

	"github.com/google/uuid"
)

// Story: konten sementara yang expired 24 jam setelah upload.
// visibility: [ADV] all_followers, close_friends
type Story struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	MediaURL      *string    `gorm:"type:varchar(500);null" json:"media_url,omitempty"`
	MediaType     string     `gorm:"type:varchar(20);not null;default:'text'" json:"media_type"` // image, video, text
	TextContent   *string    `gorm:"type:text;null" json:"text_content,omitempty"`
	MusicTrackURL *string    `gorm:"type:varchar(500);null" json:"music_track_url,omitempty"` // [ADV]
	Visibility    string     `gorm:"type:varchar(20);not null;default:'all_followers'" json:"visibility"` // [ADV]
	ExpiresAt     time.Time  `gorm:"not null;index" json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`

	// Relations
	User  User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Views []StoryView `gorm:"foreignKey:StoryID" json:"views,omitempty"`
	Tags  []StoryTag  `gorm:"foreignKey:StoryID" json:"tags,omitempty"` // [ADV]
}

func (Story) TableName() string {
	return "stories"
}
