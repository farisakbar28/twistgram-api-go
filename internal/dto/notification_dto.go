package dto

import "time"

type CreateNotificationRequest struct {
	RecipientID string  `json:"recipient_id" binding:"required,uuid"`
	ActorID     string  `json:"actor_id" binding:"required,uuid"`
	Type        string  `json:"type" binding:"required"`
	ReferenceID *string `json:"reference_id,omitempty"`
}

type NotificationResponse struct {
	ID          string    `json:"id"`
	RecipientID string    `json:"recipient_id"`
	ActorID     string    `json:"actor_id"`
	Type        string    `json:"type"`
	ReferenceID *string   `json:"reference_id,omitempty"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}
