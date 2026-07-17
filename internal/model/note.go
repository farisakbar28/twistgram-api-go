package model

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID           uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID       uuid.UUID   `gorm:"type:uuid;not null"`
	Content      string      `gorm:"type:varchar(60);not null"`
	AudioTrackID *uuid.UUID  `gorm:"type:uuid"`
	ExpiresAt    time.Time   `gorm:"not null"`
	CreatedAt    time.Time   `gorm:"autoCreateTime"`
	
	User         User        `gorm:"foreignKey:UserID"`
	AudioTrack   *AudioTrack `gorm:"foreignKey:AudioTrackID"`
}
