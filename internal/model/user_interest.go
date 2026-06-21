package model

import (
	"time"

	"github.com/google/uuid"
)

type UserInterest struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	InterestCategory string    `gorm:"type:varchar(100);not null" json:"interest_category"`
	CreatedAt        time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserInterest) TableName() string {
	return "user_interests"
}
