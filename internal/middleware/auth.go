package middleware

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"twistgram-api-go/pkg/response"
)

const (
	contextUserIDKey    = "user_id"
	contextUserEmailKey = "user_email"
)

// AuthRequired adalah middleware Gin yang memvalidasi JWT dari header
// Authorization: Bearer <token> menggunakan SUPABASE_JWT_SECRET.
// Token yang valid akan mengekstrak klaim 'sub' sebagai user_id
// dan menyimpannya ke context untuk digunakan handler selanjutnya.
func AuthRequired() gin.HandlerFunc {
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
				log.Printf("unexpected JWT signing method: %v", token.Header["alg"])
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || token == nil || !token.Valid {
			if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
				log.Printf("JWT validation error: %v", err)
			}
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

		c.Set(contextUserIDKey, parsedUserID.String())

		if email, ok := claims["email"].(string); ok && strings.TrimSpace(email) != "" {
			c.Set(contextUserEmailKey, strings.TrimSpace(email))
		}

		c.Next()
	}
}

// GetUserID mengambil user_id dari context yang sudah disimpan oleh middleware AuthRequired.
// Returns empty string jika tidak ada atau tipe datanya tidak sesuai.
func GetUserID(c *gin.Context) string {
	if c == nil {
		return ""
	}

	userID, exists := c.Get(contextUserIDKey)
	if !exists {
		return ""
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return ""
	}

	return userIDStr
}

// GetUserEmail mengambil email dari context (opsional, tidak selalu ada).
func GetUserEmail(c *gin.Context) string {
	if c == nil {
		return ""
	}

	email, exists := c.Get(contextUserEmailKey)
	if !exists {
		return ""
	}

	emailStr, ok := email.(string)
	if !ok {
		return ""
	}

	return emailStr
}
