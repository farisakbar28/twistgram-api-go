package model

import (
	"time"

	"github.com/google/uuid"
)

// SearchHistory menyimpan riwayat pencarian pengguna. [ADV]
type SearchHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_search_user_query" json:"user_id"`
	Query     string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_search_user_query" json:"query"`
	QueryType string    `gorm:"type:varchar(20);not null" json:"query_type"` // user, hashtag
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (SearchHistory) TableName() string {
	return "search_history"
}
