package model

import (
	"time"

	"github.com/google/uuid"
)

// SavedPost menyimpan post yang disimpan pengguna ke koleksi pribadi.
// collection_name: [ADV] untuk mengelompokkan saved post ke folder bernama.
type SavedPost struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_saved_unique" json:"user_id"`
	PostID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_saved_unique" json:"post_id"`
	CollectionName string    `gorm:"type:varchar(100);default:'All'" json:"collection_name"` // [ADV]
	CreatedAt      time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

func (SavedPost) TableName() string {
	return "saved_posts"
}
