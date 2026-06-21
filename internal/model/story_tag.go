package model

import (
	"github.com/google/uuid"
)

// StoryTag menyimpan mention/tag pengguna pada story. [ADV]
type StoryTag struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoryID      uuid.UUID `gorm:"type:uuid;not null;index" json:"story_id"`
	TaggedUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"tagged_user_id"`

	// Relations
	Story       Story `gorm:"foreignKey:StoryID" json:"story,omitempty"`
	TaggedUser  User  `gorm:"foreignKey:TaggedUserID" json:"tagged_user,omitempty"`
}

func (StoryTag) TableName() string {
	return "story_tags"
}
