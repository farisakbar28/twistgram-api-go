package model

import (
	"time"

	"github.com/google/uuid"
)

// Like bersifat polymorphic untuk post dan comment.
type Like struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_like_unique" json:"user_id"`
	LikeableType string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_like_unique" json:"likeable_type"` // post, comment
	LikeableID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_like_unique" json:"likeable_id"`
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Like) TableName() string {
	return "likes"
}
