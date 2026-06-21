package dto

import "time"

type SocialUserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type FollowStatusResponse struct {
	TargetID string `json:"target_id"`
	Status   string `json:"status"`
}

type BlockStatusResponse struct {
	TargetID string `json:"target_id"`
	Blocked  bool   `json:"blocked"`
}

type FollowRequestResponse struct {
	ID          string             `json:"id"`
	Requester   SocialUserResponse `json:"requester"`
	RequestedAt time.Time          `json:"requested_at"`
}

type ReportRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   string `json:"target_id" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
}

type ReportResponse struct {
	ID         string    `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
