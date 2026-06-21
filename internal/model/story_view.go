package model

import (
	"time"

	"github.com/google/uuid"
)

// StoryView mencatat siapa saja yang melihat story.
type StoryView struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoryID   uuid.UUID `gorm:"type:uuid;not null;index" json:"story_id"`
	ViewerID  uuid.UUID `gorm:"type:uuid;not null;index" json:"viewer_id"`
	ViewedAt  time.Time `json:"viewed_at"`

	// Relations
	Story  Story `gorm:"foreignKey:StoryID" json:"story,omitempty"`
	Viewer User  `gorm:"foreignKey:ViewerID" json:"viewer,omitempty"`
}

func (StoryView) TableName() string {
	return "story_views"
}
