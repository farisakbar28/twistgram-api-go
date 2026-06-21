package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/pkg/response"
)

// UserHandler menangani request terkait user/profile.
type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetMe mengembalikan profil user yang sedang login (berdasarkan token JWT).
// Endpoint: GET /api/v1/users/me (PROTECTED)
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var user model.User
	if err := config.GetDB().First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "User not found")
			return
		}

		response.InternalError(c, "Failed to get user")
		return
	}

	response.Success(c, gin.H{
		"user": gin.H{
			"id":             user.ID,
			"name":           user.Name,
			"username":       user.Username,
			"email":          user.Email,
			"phone":          user.Phone,
			"phone_verified": user.PhoneVerified,
			"email_verified": user.EmailVerified,
			"bio":            user.Bio,
			"avatar_url":     user.AvatarURL,
			"is_private":     user.IsPrivate,
			"created_at":     user.CreatedAt,
			"updated_at":     user.UpdatedAt,
		},
	})
}
