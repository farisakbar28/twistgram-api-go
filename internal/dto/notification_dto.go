package dto

type CreateNotificationRequest struct {
	RecipientID string  `json:"recipient_id" binding:"required,uuid"`
	ActorID     string  `json:"actor_id" binding:"required,uuid"`
	Type        string  `json:"type" binding:"required,oneof=like comment mention follow"`
	ReferenceID *string `json:"reference_id,omitempty"`
}
