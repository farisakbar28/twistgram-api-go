package model

import (
	"time"

	"github.com/google/uuid"
)

// Comment mendukung nested reply (parent_comment_id) dan soft delete.
type Comment struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"post_id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ParentCommentID *uuid.UUID `gorm:"type:uuid;null;index" json:"parent_comment_id,omitempty"`
	Content         string     `gorm:"type:text;not null" json:"content"`
	IsPinned        bool       `gorm:"default:false" json:"is_pinned"` // [ADV]
	CreatedAt       time.Time  `json:"created_at"`
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Post    Post      `gorm:"foreignKey:PostID" json:"post,omitempty"`
	User    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentCommentID" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentCommentID" json:"replies,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}
