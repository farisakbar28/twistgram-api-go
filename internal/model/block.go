package model

import (
	"time"

	"github.com/google/uuid"
)

// Block menyimpan relasi block antar pengguna (mutual: jika A block B, B tidak bisa melihat A).
type Block struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BlockerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_block_pair" json:"blocker_id"`
	BlockedID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_block_pair" json:"blocked_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	Blocker User `gorm:"foreignKey:BlockerID" json:"blocker,omitempty"`
	Blocked User `gorm:"foreignKey:BlockedID" json:"blocked,omitempty"`
}

func (Block) TableName() string {
	return "blocks"
}
