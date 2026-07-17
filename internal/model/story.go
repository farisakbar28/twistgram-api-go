package model

import (
	"time"

	"github.com/google/uuid"
)

type Story struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	MediaURL           *string   `gorm:"type:varchar(500);null" json:"media_url,omitempty"`
	MediaType          string    `gorm:"type:varchar(20);not null;default:'text'" json:"media_type"`
	TextContent        *string   `gorm:"type:text;null" json:"text_content,omitempty"`
	MusicTrackURL      *string   `gorm:"type:varchar(500);null" json:"music_track_url,omitempty"`
	IsCloseFriendsOnly bool      `gorm:"default:false" json:"is_close_friends_only"`
	ExpiresAt          time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt          time.Time `json:"created_at"`

	User  User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Views []StoryView `gorm:"foreignKey:StoryID" json:"views,omitempty"`
	Tags  []StoryTag  `gorm:"foreignKey:StoryID" json:"tags,omitempty"`
}

func (Story) TableName() string {
	return "stories"
}
