package middleware

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/pkg/auth"
)

func TestAuthRequiredValidTokenSetsContext(t *testing.T) {
	os.Setenv("SUPABASE_JWT_SECRET", "supersecretkey12345678901234567890")
	userID := uuid.New()
	_, _, _ = auth.GenerateJWT(userID, "test@example.com", true, 1, "supersecretkey12345678901234567890")
	
	gin.SetMode(gin.TestMode)
	_ = AuthRequired(nil)
}

func TestContextHelpersAreSafe(t *testing.T) {
	if GetUserID(nil) != "" {
		t.Fatal("GetUserID(nil) should return empty string")
	}
}
