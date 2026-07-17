package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/constants"
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
		&model.AuthOTP{},
		&model.SearchHistory{},
	)
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}
	log.Println("Migration completed successfully")

	// Setup Gin router
	r := gin.Default()

	// Request ID middleware (first in chain)
	r.Use(middleware.RequestID())

	// Use security headers
	r.Use(middleware.SecurityHeaders())

	// Use CORS middleware
	corsConfig := cors.DefaultConfig()
	if cfg.CORSAllowOrigins == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = strings.Split(cfg.CORSAllowOrigins, ",")
	}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"}
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
			"status":     "ok",
			"database":   dbStatus,
			"request_id": middleware.GetRequestID(c),
			"timestamp":  time.Now().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")

	// === Initialize repositories ===
	authRepo := repository.NewAuthRepository(db)
	userRepo := repository.NewUserRepository(db)
	socialRepo := repository.NewSocialRepository(db)
	postRepo := repository.NewPostRepository(db)
	interactionRepo := repository.NewInteractionRepository(db)
	storyRepo := repository.NewStoryRepository(db)
	searchRepo := repository.NewSearchRepository(db)
	dmRepo := repository.NewDMRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	highlightRepo := repository.NewHighlightRepository(db)
	searchHistoryRepo := repository.NewSearchHistoryRepository(db)

	// === Initialize services ===
	authService := service.NewAuthService(authRepo, cfg)
	userService := service.NewUserService(userRepo)
	socialService := service.NewSocialService(socialRepo)
	postService := service.NewPostService(postRepo)
	interactionService := service.NewInteractionService(interactionRepo)
	storyService := service.NewStoryService(storyRepo)
	searchService := service.NewSearchService(searchRepo)
	dmService := service.NewDMService(dmRepo)
	notifService := service.NewNotificationService(notifRepo)
	highlightService := service.NewHighlightService(highlightRepo)
	searchHistoryService := service.NewSearchHistoryService(searchHistoryRepo)

	// === Initialize handlers (DI pattern) ===
	authHandler := handler.NewAuthHandlerWithService(authService)
	userHandler := handler.NewUserHandlerWithService(userService)
	socialHandler := handler.NewSocialHandlerWithService(socialService)
	postHandler := handler.NewPostHandlerWithService(postService)
	interactionHandler := handler.NewInteractionHandlerWithService(interactionService)
	storyHandler := handler.NewStoryHandlerWithService(storyService)
	searchHandler := handler.NewSearchHandlerWithService(searchService)
	dmHandler := handler.NewDMHandlerWithService(dmService)
	notifHandler := handler.NewNotificationHandlerWithService(notifService)
	highlightHandler := handler.NewHighlightHandlerWithService(highlightService)
	searchHistoryHandler := handler.NewSearchHistoryHandlerWithService(searchHistoryService)

	// Public routes (no auth required)
	public := v1.Group("")
	public.Use(middleware.RateLimit(constants.RateLimitPublic, constants.RateBurstPublic))
	{
		public.POST("/auth/register", authHandler.Register)
		public.GET("/auth/check-availability", authHandler.CheckAvailability)
		public.POST("/auth/verify-otp", authHandler.VerifyOTP)
		public.POST("/auth/resend-otp", authHandler.ResendOTP)
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/refresh-token", authHandler.RefreshToken)
		public.POST("/auth/logout", authHandler.Logout)
		public.POST("/auth/forgot-password", authHandler.ForgotPassword)
		public.POST("/auth/recover-username", authHandler.RecoverUsername)
		public.POST("/auth/recover-email", authHandler.RecoverEmail)
		public.POST("/auth/recover-email/complete", authHandler.CompleteRecoverEmail)
		public.POST("/auth/reset-password", authHandler.ResetPassword)
	}

	// Protected routes (auth required)
	auth := v1.Group("")
	auth.Use(middleware.AuthRequired())
	{
		// Users & Profile (readable by unverified)
		auth.GET("/users/me", userHandler.GetMe)
		auth.PATCH("/users/me", userHandler.UpdateMe)
		auth.DELETE("/users/me", userHandler.DeleteAccount)
		auth.PATCH("/users/me/privacy", userHandler.UpdatePrivacy)
		auth.GET("/users/me/interests", userHandler.GetInterests)
		auth.PUT("/users/me/interests", userHandler.SetInterests)
		auth.GET("/users/:identifier", userHandler.GetByUsername)

		// Read-only routes (accessible to unverified users)
		auth.GET("/feed", postHandler.Feed)
		auth.GET("/posts/:id", postHandler.GetByID)
		auth.GET("/users/me/posts", postHandler.MyPosts)
		auth.GET("/users/:identifier/posts", postHandler.UserPosts)
		auth.GET("/users/:identifier/followers", socialHandler.Followers)
		auth.GET("/users/:identifier/following", socialHandler.Following)
		auth.GET("/users/me/follow-requests", socialHandler.FollowRequests)
		auth.GET("/users/me/close-friends", socialHandler.ListCloseFriends)
		auth.GET("/users/me/blocked", socialHandler.GetBlockedUsers)
		auth.GET("/posts/:id/comments", interactionHandler.ListComments)
		auth.GET("/users/me/saved", interactionHandler.ListSavedPosts)
		auth.GET("/stories/feed", storyHandler.Feed)
		auth.GET("/stories/:id", storyHandler.GetByID)
		auth.GET("/stories/:id/viewers", storyHandler.Viewers)
		auth.GET("/highlights", highlightHandler.List)
		auth.GET("/search", searchHandler.Search)
		auth.GET("/hashtags/:tag/posts", searchHandler.HashtagPosts)
		auth.GET("/search/history", searchHistoryHandler.List)
		auth.GET("/conversations", dmHandler.ListConversations)
		auth.GET("/conversations/requests", dmHandler.ListMessageRequests)
		auth.GET("/conversations/:id/messages", dmHandler.ListMessages)
		auth.GET("/notifications", notifHandler.List)

		// AUTH-02: Write operations require email verification
		verified := v1.Group("")
		verified.Use(middleware.AuthRequired(), middleware.EmailVerifiedRequired())
		{
			// Follow & Social
			verified.POST("/users/:identifier/follow", socialHandler.Follow)
			verified.DELETE("/users/:identifier/follow", socialHandler.Unfollow)
			verified.DELETE("/users/:identifier/followers", socialHandler.RemoveFollower)
			verified.POST("/users/:identifier/follow-requests/approve", socialHandler.ApproveFollowRequest)
			verified.POST("/users/:identifier/follow-requests/decline", socialHandler.DeclineFollowRequest)

			// Close Friends [ADV]
			verified.POST("/users/:identifier/close-friends", socialHandler.AddCloseFriend)
			verified.DELETE("/users/:identifier/close-friends", socialHandler.RemoveCloseFriend)

			// Block & Report
			verified.POST("/users/:identifier/block", socialHandler.Block)
			verified.DELETE("/users/:identifier/block", socialHandler.Unblock)
			verified.POST("/reports", socialHandler.Report)

			// Posts (write)
			verified.POST("/posts", postHandler.Create)
			verified.PATCH("/posts/:id", postHandler.EditCaption)
			verified.DELETE("/posts/:id", postHandler.Delete)
			verified.POST("/posts/:id/archive", postHandler.Archive)
			verified.POST("/posts/:id/unarchive", postHandler.Unarchive)
			verified.DELETE("/posts/:id/tags/:taggedUserId", postHandler.RemoveTag)

			// Interactions (write)
			verified.POST("/posts/:id/like", interactionHandler.LikePost)
			verified.DELETE("/posts/:id/like", interactionHandler.UnlikePost)
			verified.POST("/posts/:id/comments", interactionHandler.Comment)
			verified.DELETE("/posts/:id/comments/:comment_id", interactionHandler.DeleteComment)
			verified.POST("/posts/:id/comments/:comment_id/like", interactionHandler.LikeComment)
			verified.POST("/posts/:id/save", interactionHandler.SavePost)
			verified.DELETE("/posts/:id/save", interactionHandler.UnsavePost)
			verified.POST("/posts/:id/share", interactionHandler.SharePost)

			// Stories (write)
			verified.POST("/stories", storyHandler.Create)
			verified.DELETE("/stories/:id", storyHandler.Delete)
			verified.POST("/stories/:id/views", storyHandler.RecordView)

			// Story Highlights [ADV] (write)
			verified.POST("/highlights", highlightHandler.Create)
			verified.PATCH("/highlights/:id", highlightHandler.Update)
			verified.DELETE("/highlights/:id", highlightHandler.Delete)
			verified.POST("/highlights/:id/stories", highlightHandler.AddStory)
			verified.DELETE("/highlights/:id/stories/:story_id", highlightHandler.RemoveStory)

			// Search History [ADV] (write)
			verified.POST("/search/history", searchHistoryHandler.Save)
			verified.DELETE("/search/history/:id", searchHistoryHandler.DeleteItem)
			verified.DELETE("/search/history", searchHistoryHandler.DeleteAll)

			// Direct Messages (write)
			verified.POST("/conversations", dmHandler.StartConversation)
			verified.POST("/conversations/:id/messages", dmHandler.SendMessage)

			// Notifications (write)
			verified.POST("/notifications/read-all", notifHandler.ReadAll)
			verified.POST("/notifications", notifHandler.Create)
			verified.POST("/notifications/:id/read", notifHandler.Read)
		}
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
