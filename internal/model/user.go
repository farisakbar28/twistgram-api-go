package model

import (
	"time"

	"github.com/google/uuid"
)

// User menyimpan data profil pengguna, sinkron dengan auth.users Supabase.
// id mengikuti UUID dari Supabase Auth (foreign key ke auth.users).
type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name           string     `gorm:"type:varchar(255);not null" json:"name"`
	Username       string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"username"`
	Email          string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone          *string    `gorm:"type:varchar(50);null" json:"phone,omitempty"`
	PhoneVerified  bool       `gorm:"default:false" json:"phone_verified"`
	EmailVerified  bool       `gorm:"default:false" json:"email_verified"`
	Bio            *string    `gorm:"type:text;null" json:"bio,omitempty"`
	AvatarURL      *string    `gorm:"type:varchar(500);null" json:"avatar_url,omitempty"`
	IsPrivate      bool       `gorm:"default:false" json:"is_private"`
	LastUsernameAt *time.Time `gorm:"null" json:"last_username_at,omitempty"` // SOC-05: 1x change per month
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Posts             []Post             `gorm:"foreignKey:UserID" json:"posts,omitempty"`
	Stories           []Story            `gorm:"foreignKey:UserID" json:"stories,omitempty"`
	Comments          []Comment          `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Notifications     []Notification     `gorm:"foreignKey:RecipientID" json:"notifications,omitempty"`
	SentNotifications []Notification     `gorm:"foreignKey:ActorID" json:"sent_notifications,omitempty"`
	Conversations     []ConversationParticipant `gorm:"foreignKey:UserID" json:"conversations,omitempty"`
	Messages          []Message          `gorm:"foreignKey:SenderID" json:"messages,omitempty"`
	Reports           []Report           `gorm:"foreignKey:ReporterID" json:"reports,omitempty"`
}

func (User) TableName() string {
	return "users"
}
