package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const contextRequestIDKey = "request_id"

// RequestID adds a unique request ID to every request for tracing.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set(contextRequestIDKey, rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}

// GetRequestID retrieves the request ID from gin context.
func GetRequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	val, exists := c.Get(contextRequestIDKey)
	if !exists {
		return ""
	}
	rid, ok := val.(string)
	if !ok {
		return ""
	}
	return rid
}
