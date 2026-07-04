package model

import (
	"time"

	"github.com/google/uuid"
)

type AuthOTP struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Code      string    `gorm:"type:varchar(6);not null" json:"code"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"` // "signup", "recovery", dll
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

func (AuthOTP) TableName() string {
	return "auth_otps"
}
