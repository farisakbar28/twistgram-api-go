package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandlerWithService(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	profile, err := h.userService.GetMe(c.Request.Context(), userID)
	h.handleServiceResult(c, profile, err)
}

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

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	h.handleServiceResult(c, profile, err)
}

func (h *UserHandler) GetInterests(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	res, err := h.userService.GetInterests(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "User not found")
		} else {
			response.InternalError(c, "Failed to fetch interests")
		}
		return
	}
	response.Success(c, res)
}

func (h *UserHandler) SetInterests(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	var req dto.UserInterestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.userService.SetInterests(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "User not found")
		} else {
			response.InternalError(c, "Failed to update interests")
		}
		return
	}
	response.Success(c, res)
}

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

	profile, err := h.userService.UpdatePrivacy(c.Request.Context(), userID, req)
	h.handleServiceResult(c, profile, err)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.userService.DeleteAccount(c.Request.Context(), userID, req.Password); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		case errors.Is(err, service.ErrWrongPassword):
			response.BadRequest(c, "Incorrect password")
		default:
			response.InternalError(c, "Failed to delete account")
		}
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

func (h *UserHandler) GetByUsername(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	profile, err := h.userService.GetProfileByUsername(c.Request.Context(), c.Param("identifier"), viewerID)
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
