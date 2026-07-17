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
	// Primary field names (backend convention)
	Email string `json:"email"`
	Token string `json:"token"`
	Type  string `json:"type"`

	// Alternate field names (frontend sends these)
	Otp      string `json:"otp"`
	Purpose  string `json:"purpose"`
	Identity string `json:"identifier"`
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required"`
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

type CompleteRecoverEmailRequest struct {
	Username string `json:"username" binding:"required"`
	Token    string `json:"token" binding:"required"`
	NewEmail string `json:"new_email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthSessionResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

type AuthUserResponse struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    EmailVerified bool   `json:"email_verified"`
}

type AuthResponse struct {
	User    AuthUserResponse     `json:"user"`
	Session *AuthSessionResponse `json:"session,omitempty"`
	Message string               `json:"message,omitempty"`
}
