package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandlerWithService(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.authService.Register(c.Request.Context(), req)
	if h.handleAuthError(c, err) {
		return
	}
	response.Created(c, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.authService.Login(c.Request.Context(), req)
	if h.handleAuthError(c, err) {
		return
	}
	response.Success(c, res)
}
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.authService.VerifyOTP(c.Request.Context(), req)
	if h.handleAuthError(c, err) {
		return
	}
	response.Success(c, res)
}

func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req dto.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleAuthError(c, h.authService.ResendOTP(c.Request.Context(), req)) {
		return
	}
	response.Success(c, gin.H{"message": "OTP sent"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleAuthError(c, h.authService.ForgotPassword(c.Request.Context(), req)) {
		return
	}
	response.Success(c, gin.H{"message": "OTP sent"})
}

func (h *AuthHandler) RecoverUsername(c *gin.Context) {
	var req dto.RecoverUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleAuthError(c, h.authService.RecoverUsername(c.Request.Context(), req)) {
		return
	}
	response.Success(c, gin.H{"message": "OTP sent"})
}

func (h *AuthHandler) RecoverEmail(c *gin.Context) {
	var req dto.RecoverEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleAuthError(c, h.authService.RecoverEmail(c.Request.Context(), req)) {
		return
	}
	response.Success(c, gin.H{"message": "Recovery process initiated"})
}

func (h *AuthHandler) CompleteRecoverEmail(c *gin.Context) {
	var req dto.CompleteRecoverEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleAuthError(c, h.authService.CompleteRecoverEmail(c.Request.Context(), req)) {
		return
	}
	response.Success(c, gin.H{"message": "Email updated successfully"})
}

func (h *AuthHandler) CheckAvailability(c *gin.Context) {
	var req dto.CheckAvailabilityRequest
	_ = c.ShouldBindQuery(&req)
	res, err := h.authService.CheckAvailability(c.Request.Context(), req)
	if h.handleAuthError(c, err) {
		return
	}
	if res.Available {
		response.Success(c, res)
	} else {
		response.Error(c, 409, res.Message)
	}
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleAuthError(c, h.authService.ResetPassword(c.Request.Context(), req)) {
		return
	}
	response.Success(c, gin.H{"message": "Password updated"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.authService.RefreshToken(c.Request.Context(), req)
	if h.handleAuthError(c, err) {
		return
	}
	response.Success(c, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error) bool {

	if err == nil {
		return false
	}
	if errors.Is(err, service.ErrInvalidInput) {
		response.BadRequest(c, "Invalid request data")
	} else if appErr, ok := err.(*service.AppError); ok {
		response.Error(c, appErr.Code, appErr.Message)
	} else {
		response.InternalError(c, "Internal server error")
	}
	return true
}
