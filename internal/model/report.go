package model

import (
	"time"

	"github.com/google/uuid"
)

// Report menyimpan laporan pengguna terhadap konten atau pengguna lain.
// target_type: user, post, comment
// reason: spam, inappropriate, harassment, fake_account, other
// status: pending, reviewed, action_taken, dismissed
type Report struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReporterID uuid.UUID `gorm:"type:uuid;not null;index" json:"reporter_id"`
	TargetType string    `gorm:"type:varchar(20);not null" json:"target_type"` // user, post, comment
	TargetID   uuid.UUID `gorm:"type:uuid;not null" json:"target_id"`
	Reason     string    `gorm:"type:varchar(30);not null" json:"reason"`
	Status     string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`

	// Relations
	Reporter User `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
}

func (Report) TableName() string {
	return "reports"
}
