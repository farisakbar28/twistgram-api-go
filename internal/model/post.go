package model

import (
	"time"

	"github.com/google/uuid"
)

// Post menyimpan data konten post pengguna. Soft delete via deleted_at.
type Post struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Caption          *string    `gorm:"type:text;null" json:"caption,omitempty"`
	Location         *string    `gorm:"type:varchar(255);null" json:"location,omitempty"` // [ADV]
	IsArchived       bool       `gorm:"default:false" json:"is_archived"`
	CommentsDisabled bool       `gorm:"default:false" json:"comments_disabled"` // [ADV]
	CreatedAt        time.Time  `json:"created_at"`
	DeletedAt        *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Relations
	User     User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Media    []PostMedia `gorm:"foreignKey:PostID" json:"media,omitempty"`
	Tags     []PostTag   `gorm:"foreignKey:PostID" json:"tags,omitempty"`
	Hashtags []PostHashtag `gorm:"foreignKey:PostID" json:"hashtags,omitempty"`
	Comments []Comment   `gorm:"foreignKey:PostID" json:"comments,omitempty"`
}

func (Post) TableName() string {
	return "posts"
}
