package model

import (
	"github.com/google/uuid"
)

// PostTag menyimpan mention/tag pengguna pada post.
type PostTag struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID       uuid.UUID `gorm:"type:uuid;not null;index" json:"post_id"`
	TaggedUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"tagged_user_id"`
	PositionX    *float64  `gorm:"null" json:"position_x,omitempty"` // [ADV] koordinat tag
	PositionY    *float64  `gorm:"null" json:"position_y,omitempty"` // [ADV] koordinat tag

	// Relations
	Post        Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
	TaggedUser  User `gorm:"foreignKey:TaggedUserID" json:"tagged_user,omitempty"`
}

func (PostTag) TableName() string {
	return "post_tags"
}
