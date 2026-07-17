package model

import (
	"time"

	"github.com/google/uuid"
)

type AudioTrack struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Title         string     `gorm:"type:varchar(150);not null"`
	Artist        string     `gorm:"type:varchar(100);not null"`
	URL           string     `gorm:"type:text;not null"`
	Duration      int        `gorm:"not null"`
	IsOriginal    bool       `gorm:"default:false"`
	UploaderID    *uuid.UUID `gorm:"type:uuid"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
}
