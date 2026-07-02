package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type VerifyOTPRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Token  string `json:"token" binding:"required"`
	Type   string `json:"type" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RecoverUsernameRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RecoverEmailRequest struct {
	Username string `json:"username" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthSessionResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

type AuthUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type AuthResponse struct {
	User    AuthUserResponse    `json:"user"`
	Session *AuthSessionResponse `json:"session,omitempty"`
	Message string              `json:"message,omitempty"`
}
