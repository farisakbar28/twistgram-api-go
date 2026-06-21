package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret"

func TestAuthRequiredValidTokenSetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("SUPABASE_JWT_SECRET", testJWTSecret)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	token := makeTestToken(t, jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"email": "user@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	router := gin.New()
	router.GET("/protected", AuthRequired(), func(c *gin.Context) {
		if got := GetUserID(c); got != userID {
			t.Fatalf("expected user_id %q, got %q", userID, got)
		}
		if got := GetUserEmail(c); got != "user@example.com" {
			t.Fatalf("expected user_email %q, got %q", "user@example.com", got)
		}
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d body %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

func TestAuthRequiredRejectsInvalidTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		secret    string
		header    string
		makeToken func(t *testing.T) string
	}{
		{
			name:   "missing secret fails closed",
			secret: "",
			header: "Bearer anything",
		},
		{
			name:   "missing bearer token",
			secret: testJWTSecret,
			header: "",
		},
		{
			name:   "expired token",
			secret: testJWTSecret,
			makeToken: func(t *testing.T) string {
				return makeTestToken(t, jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "550e8400-e29b-41d4-a716-446655440000",
					"exp": time.Now().Add(-time.Hour).Unix(),
				})
			},
		},
		{
			name:   "non uuid subject",
			secret: testJWTSecret,
			makeToken: func(t *testing.T) string {
				return makeTestToken(t, jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "not-a-uuid",
					"exp": time.Now().Add(time.Hour).Unix(),
				})
			},
		},
		{
			name:   "wrong signing method",
			secret: testJWTSecret,
			makeToken: func(t *testing.T) string {
				return makeTestToken(t, jwt.SigningMethodHS384, jwt.MapClaims{
					"sub": "550e8400-e29b-41d4-a716-446655440000",
					"exp": time.Now().Add(time.Hour).Unix(),
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnvForTest(t, "SUPABASE_JWT_SECRET", tt.secret)

			router := gin.New()
			router.GET("/protected", AuthRequired(), func(c *gin.Context) {
				t.Fatal("handler should not be called")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			header := tt.header
			if tt.makeToken != nil {
				header = "Bearer " + tt.makeToken(t)
			}
			if header != "" {
				req.Header.Set("Authorization", header)
			}

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d body %s", http.StatusUnauthorized, w.Code, w.Body.String())
			}
		})
	}
}

func TestContextHelpersAreSafe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	if got := GetUserID(nil); got != "" {
		t.Fatalf("expected empty user_id for nil context, got %q", got)
	}
	if got := GetUserEmail(nil); got != "" {
		t.Fatalf("expected empty user_email for nil context, got %q", got)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(contextUserIDKey, 123)
	c.Set(contextUserEmailKey, 456)

	if got := GetUserID(c); got != "" {
		t.Fatalf("expected empty user_id for non-string context value, got %q", got)
	}
	if got := GetUserEmail(c); got != "" {
		t.Fatalf("expected empty user_email for non-string context value, got %q", got)
	}
}

func makeTestToken(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()

	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	return token
}

func setEnvForTest(t *testing.T, key, value string) {
	t.Helper()

	original, hadOriginal := os.LookupEnv(key)
	if value == "" {
		_ = os.Unsetenv(key)
	} else {
		_ = os.Setenv(key, value)
	}

	t.Cleanup(func() {
		if hadOriginal {
			_ = os.Setenv(key, original)
			return
		}
		_ = os.Unsetenv(key)
	})
}
