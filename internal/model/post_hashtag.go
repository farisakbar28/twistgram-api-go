package model

import (
	"github.com/google/uuid"
)

// PostHashtag adalah tabel relasi many-to-many antara post dan hashtag.
type PostHashtag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_post_hashtag" json:"post_id"`
	HashtagID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_post_hashtag" json:"hashtag_id"`

	// Relations
	Post    Post    `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Hashtag Hashtag `gorm:"foreignKey:HashtagID" json:"hashtag,omitempty"`
}

func (PostHashtag) TableName() string {
	return "post_hashtags"
}
