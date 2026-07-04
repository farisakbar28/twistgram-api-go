package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/handler"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/internal/service"
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

	// Use security headers
	r.Use(middleware.SecurityHeaders())

	// Use CORS middleware
	corsConfig := cors.DefaultConfig()
	if cfg.CORSAllowOrigins == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = strings.Split(cfg.CORSAllowOrigins, ",")
	}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	// Limit JSON body size to 2MB (2 * 1024 * 1024)
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
		c.Next()
	})

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
	authRepo := repository.NewAuthRepository(db, cfg.SupabaseURL, cfg.SupabaseAnonKey)
	authHandler := handler.NewAuthHandlerWithService(service.NewAuthService(authRepo))
	public := v1.Group("")
	// Apply rate limiting to public endpoints (5 req/sec, burst 10)
	public.Use(middleware.RateLimit(5, 10))
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/verify-otp", authHandler.VerifyOTP)
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/forgot-password", authHandler.ForgotPassword)
		public.POST("/auth/recover-username", authHandler.RecoverUsername)
		public.POST("/auth/recover-email", authHandler.RecoverEmail)
		public.POST("/auth/reset-password", authHandler.ResetPassword)
	}

	// Protected routes (auth required)
	userHandler := handler.NewUserHandler()
	socialHandler := handler.NewSocialHandler()
	postHandler := handler.NewPostHandler()
	interactionHandler := handler.NewInteractionHandler()
	storyHandler := handler.NewStoryHandler()
	searchHandler := handler.NewSearchHandler()
	dmHandler := handler.NewDMHandler()
	notificationHandler := handler.NewNotificationHandler()
	auth := v1.Group("")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/users/me", userHandler.GetMe)
		auth.PATCH("/users/me", userHandler.UpdateMe)
		auth.PATCH("/users/me/privacy", userHandler.UpdatePrivacy)
		auth.GET("/users/me/interests", userHandler.GetInterests)
		auth.PUT("/users/me/interests", userHandler.SetInterests)
		auth.GET("/users/:identifier", userHandler.GetByUsername)
		auth.POST("/users/:identifier/follow", socialHandler.Follow)
		auth.DELETE("/users/:identifier/follow", socialHandler.Unfollow)
		auth.GET("/users/:identifier/followers", socialHandler.Followers)
		auth.GET("/users/:identifier/following", socialHandler.Following)
		auth.DELETE("/users/:identifier/followers", socialHandler.RemoveFollower)
		auth.GET("/users/me/follow-requests", socialHandler.FollowRequests)
		auth.POST("/users/:identifier/follow-requests/approve", socialHandler.ApproveFollowRequest)
		auth.POST("/users/:identifier/follow-requests/decline", socialHandler.DeclineFollowRequest)
		auth.POST("/users/:identifier/block", socialHandler.Block)
		auth.DELETE("/users/:identifier/block", socialHandler.Unblock)
		auth.POST("/reports", socialHandler.Report)
		auth.POST("/posts", postHandler.Create)
		auth.PATCH("/posts/:id", postHandler.EditCaption)
		auth.GET("/feed", postHandler.Feed)
		auth.GET("/users/me/posts", postHandler.MyPosts)
		auth.DELETE("/posts/:id", postHandler.Delete)
		auth.POST("/posts/:id/archive", postHandler.Archive)
		auth.POST("/posts/:id/unarchive", postHandler.Unarchive)
		auth.POST("/posts/:id/like", interactionHandler.LikePost)
		auth.DELETE("/posts/:id/like", interactionHandler.UnlikePost)
		auth.GET("/posts/:id/comments", interactionHandler.ListComments)
		auth.POST("/posts/:id/comments", interactionHandler.Comment)
		auth.DELETE("/posts/:id/comments/:comment_id", interactionHandler.DeleteComment)
		auth.POST("/posts/:id/comments/:comment_id/like", interactionHandler.LikeComment)
		auth.GET("/users/me/saved", interactionHandler.ListSavedPosts)
		auth.POST("/posts/:id/save", interactionHandler.SavePost)
		auth.DELETE("/posts/:id/save", interactionHandler.UnsavePost)
		auth.POST("/posts/:id/share", interactionHandler.SharePost)
		auth.POST("/stories", storyHandler.Create)
		auth.GET("/stories/feed", storyHandler.Feed)
		auth.GET("/stories/:id", storyHandler.GetByID)
		auth.POST("/stories/:id/views", storyHandler.RecordView)
		auth.GET("/stories/:id/viewers", storyHandler.Viewers)
		auth.GET("/search", searchHandler.Search)
		auth.GET("/conversations", dmHandler.ListConversations)
		auth.POST("/conversations", dmHandler.StartConversation)
		auth.GET("/conversations/:id/messages", dmHandler.ListMessages)
		auth.POST("/conversations/:id/messages", dmHandler.SendMessage)
		auth.GET("/notifications", notificationHandler.List)
		auth.POST("/notifications/:id/read", notificationHandler.Read)
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
