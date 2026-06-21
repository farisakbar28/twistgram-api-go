package model

import (
	"github.com/google/uuid"
)

// HighlightStory adalah tabel relasi many-to-many antara highlight dan story. [ADV]
type HighlightStory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HighlightID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_highlight_story" json:"highlight_id"`
	StoryID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_highlight_story" json:"story_id"`

	// Relations
	Highlight Highlight `gorm:"foreignKey:HighlightID" json:"highlight,omitempty"`
	Story     Story     `gorm:"foreignKey:StoryID" json:"story,omitempty"`
}

func (HighlightStory) TableName() string {
	return "highlight_stories"
}
