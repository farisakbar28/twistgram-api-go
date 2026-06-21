package model

import (
	"time"

	"github.com/google/uuid"
)

// Hashtag menyimpan daftar hashtag unik.
type Hashtag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Tag       string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"tag"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	Posts []PostHashtag `gorm:"foreignKey:HashtagID" json:"posts,omitempty"`
}

func (Hashtag) TableName() string {
	return "hashtags"
}
