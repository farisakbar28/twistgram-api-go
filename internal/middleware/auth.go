package middleware

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/pkg/response"
)

const (
	contextUserIDKey        = "user_id"
	contextUserEmailKey     = "user_email"
	contextEmailVerifiedKey = "email_verified"
	contextTokenVersionKey  = "token_version"
)

func AuthRequired(db *gorm.DB) gin.HandlerFunc {
	jwtSecret := strings.TrimSpace(os.Getenv("SUPABASE_JWT_SECRET"))

	return func(c *gin.Context) {
		if jwtSecret == "" {
			log.Println("SUPABASE_JWT_SECRET is not configured")
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		
		if err != nil || token == nil || !token.Valid {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		userID, err := claims.GetSubject()
		if err != nil || strings.TrimSpace(userID) == "" {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// AUTH-05: Token version check
		var tokenVersion int
		if tv, ok := claims["token_version"]; ok {
			switch v := tv.(type) {
			case float64:
				tokenVersion = int(v)
			case int:
				tokenVersion = v
			}
		}

		if db != nil {
			var dbVersion int
			err := db.WithContext(context.Background()).Table("users").Select("token_version").Where("id = ?", parsedUserID).Scan(&dbVersion).Error
			if err == nil && dbVersion != tokenVersion {
				response.Unauthorized(c, "Session expired, please login again")
				c.Abort()
				return
			}
		}

		c.Set(contextUserIDKey, parsedUserID.String())
		if email, ok := claims["email"].(string); ok {
			c.Set(contextUserEmailKey, strings.TrimSpace(email))
		}
		if emailVerified, ok := claims["email_verified"].(bool); ok {
			c.Set(contextEmailVerifiedKey, emailVerified)
		}
		c.Set(contextTokenVersionKey, tokenVersion)

		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	if c == nil { return "" }
	userID, exists := c.Get(contextUserIDKey)
	if !exists { return "" }
	userIDStr, ok := userID.(string)
	if !ok { return "" }
	return userIDStr
}

func GetUserEmail(c *gin.Context) string {
	if c == nil { return "" }
	email, exists := c.Get(contextUserEmailKey)
	if !exists { return "" }
	emailStr, ok := email.(string)
	if !ok { return "" }
	return emailStr
}

func EmailVerifiedRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}
		verified, exists := c.Get(contextEmailVerifiedKey)
		if !exists {
			response.Forbidden(c, "Email verification required")
			c.Abort()
			return
		}
		isVerified, ok := verified.(bool)
		if !ok || !isVerified {
			response.Forbidden(c, "Email verification required")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetTokenVersion(c *gin.Context) int {
	if c == nil { return 0 }
	tv, exists := c.Get(contextTokenVersionKey)
	if !exists { return 0 }
	v, ok := tv.(int)
	if !ok { return 0 }
	return v
}
