package dto

type CheckAvailabilityRequest struct {
	Username string `form:"username"`
	Email    string `form:"email"`
}

type CheckAvailabilityResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}
