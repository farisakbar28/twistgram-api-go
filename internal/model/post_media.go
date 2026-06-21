package model

import (
	"github.com/google/uuid"
)

type PostMedia struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID        uuid.UUID `gorm:"type:uuid;not null;index" json:"post_id"`
	MediaURL      string    `gorm:"type:varchar(500);not null" json:"media_url"`
	MediaType     string    `gorm:"type:varchar(20);not null" json:"media_type"` // image, video
	OrderIndex    int       `gorm:"default:0" json:"order_index"`                // [ADV] carousel order
	MusicTrackURL *string   `gorm:"type:varchar(500);null" json:"music_track_url,omitempty"` // [ADV]

	// Relations
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

func (PostMedia) TableName() string {
	return "post_media"
}
