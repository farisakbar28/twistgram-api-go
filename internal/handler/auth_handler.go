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

func NewAuthHandlerWithService(authService *service.AuthService) *AuthHandler { return &AuthHandler{authService: authService} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }
	res, err := h.authService.Register(req)
	if h.handleAuthError(c, err) { return }
	response.Created(c, res)
}

func (h *AuthHandler) Login(c *gin.Context) { var req dto.LoginRequest; if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }; res, err := h.authService.Login(req); if h.handleAuthError(c, err) { return }; response.Success(c, res) }
func (h *AuthHandler) VerifyOTP(c *gin.Context) { var req dto.VerifyOTPRequest; if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }; res, err := h.authService.VerifyOTP(req); if h.handleAuthError(c, err) { return }; response.Success(c, res) }
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }
	if h.handleAuthError(c, h.authService.ForgotPassword(req)) { return }
	response.Success(c, gin.H{"message": "OTP sent"})
}

func (h *AuthHandler) RecoverUsername(c *gin.Context) {
	var req dto.RecoverUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }
	if h.handleAuthError(c, h.authService.RecoverUsername(req)) { return }
	response.Success(c, gin.H{"message": "OTP sent"})
}

func (h *AuthHandler) RecoverEmail(c *gin.Context) {
	var req dto.RecoverEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }
	if h.handleAuthError(c, h.authService.RecoverEmail(req)) { return }
	response.Success(c, gin.H{"message": "OTP sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }
	if h.handleAuthError(c, h.authService.ResetPassword(req)) { return }
	response.Success(c, gin.H{"message": "Password updated"})
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error) bool {
	if err == nil { return false }
	if errors.Is(err, service.ErrInvalidInput) { response.BadRequest(c, "Invalid request data") } else { response.InternalError(c, "Auth service unavailable") }
	return true
}
