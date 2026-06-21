package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/handler"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/pkg/response"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db := config.InitDatabase(cfg)

	// AutoMigrate all models
	log.Println("Running database migration...")
	err := db.AutoMigrate(
		&model.User{},
		&model.UserInterest{},
		&model.Follow{},
		&model.Block{},
		&model.Post{},
		&model.PostMedia{},
		&model.PostTag{},
		&model.Hashtag{},
		&model.PostHashtag{},
		&model.Like{},
		&model.Comment{},
		&model.SavedPost{},
		&model.Story{},
		&model.StoryView{},
		&model.StoryTag{},
		&model.Highlight{},
		&model.HighlightStory{},
		&model.Conversation{},
		&model.ConversationParticipant{},
		&model.Message{},
		&model.Notification{},
		&model.Report{},
	)
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}
	log.Println("Migration completed successfully")

	// Setup Gin router
	r := gin.Default()

	// Health check endpoint (public)
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

	// API v1 routes
	v1 := r.Group("/api/v1")

	// Public routes (no auth required)
	// (to be added in later phases)

	// Protected routes (auth required)
	userHandler := handler.NewUserHandler()
	auth := v1.Group("")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/users/me", userHandler.GetMe)
		auth.PATCH("/users/me", userHandler.UpdateMe)
		auth.PATCH("/users/me/privacy", userHandler.UpdatePrivacy)
		auth.GET("/users/:username", userHandler.GetByUsername)
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
