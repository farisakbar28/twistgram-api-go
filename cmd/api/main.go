package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/pkg/response"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db := config.InitDatabase(cfg)
	_ = db // Will be used in later phases

	// Setup Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := config.GetDB().DB()
		dbStatus := "connected"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "disconnected"
		}

		response.Success(c, gin.H{
			"status":    "ok",
			"database":  dbStatus,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
