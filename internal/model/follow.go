package model

import (
	"time"

	"github.com/google/uuid"
)

// Follow menyimpan relasi follow antar pengguna.
// Status: accepted (publik langsung follow) atau pending (akun privat, perlu approval).
// is_close_friend: [ADV] ditandai oleh following_id terhadap follower_id.
type Follow struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FollowerID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_follow_pair" json:"follower_id"`
	FollowingID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_follow_pair" json:"following_id"`
	Status       string    `gorm:"type:varchar(20);not null;default:'accepted'" json:"status"` // accepted, pending
	IsCloseFriend bool    `gorm:"default:false" json:"is_close_friend"` // [ADV]
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	Follower  User `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Following User `gorm:"foreignKey:FollowingID" json:"following,omitempty"`
}

func (Follow) TableName() string {
	return "follows"
}
