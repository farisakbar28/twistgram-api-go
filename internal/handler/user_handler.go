package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

// UserHandler menangani request terkait user/profile.
type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	repo := repository.NewUserRepository(config.GetDB())
	return &UserHandler{userService: service.NewUserService(repo)}
}

func NewUserHandlerWithService(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMe mengembalikan profil user yang sedang login (berdasarkan token JWT).
// Endpoint: GET /api/v1/users/me (PROTECTED)
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	profile, err := h.userService.GetMe(userID)
	h.handleServiceResult(c, profile, err)
}

// UpdateMe mengubah profil user yang sedang login.
// Endpoint: PATCH /api/v1/users/me (PROTECTED)
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	profile, err := h.userService.UpdateProfile(userID, req)
	h.handleServiceResult(c, profile, err)
}

// UpdatePrivacy mengubah status privasi user yang sedang login.
// Endpoint: PATCH /api/v1/users/me/privacy (PROTECTED)
func (h *UserHandler) UpdatePrivacy(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.UpdatePrivacyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	profile, err := h.userService.UpdatePrivacy(userID, req)
	h.handleServiceResult(c, profile, err)
}

// GetByUsername mengembalikan profil user berdasarkan username.
// Endpoint: GET /api/v1/users/:username (PROTECTED)
func (h *UserHandler) GetByUsername(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	profile, err := h.userService.GetProfileByUsername(c.Param("username"), viewerID)
	h.handleServiceResult(c, profile, err)
}

func getAuthenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return uuid.Nil, false
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		response.Unauthorized(c, "Invalid authenticated user")
		return uuid.Nil, false
	}
	return parsedUserID, true
}

func (h *UserHandler) handleServiceResult(c *gin.Context, profile *dto.UserProfileResponse, err error) {
	if err == nil {
		response.Success(c, gin.H{"user": profile})
		return
	}

	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, "Invalid request data")
	case errors.Is(err, service.ErrUserNotFound):
		response.NotFound(c, "User not found")
	case errors.Is(err, service.ErrUsernameTaken):
		response.BadRequest(c, "Username already taken")
	case errors.Is(err, service.ErrUsernameChangeLimited):
		response.BadRequest(c, "Username can only be changed once per month")
	case errors.Is(err, service.ErrUserBlocked):
		response.ForbiddenCode(c, "USER_BLOCKED", "You cannot view this profile because a block exists between these users")
	default:
		response.InternalError(c, "Failed to process user request")
	}
}
