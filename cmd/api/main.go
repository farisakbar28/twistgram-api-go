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
		// Users & Profile
		auth.GET("/users/me", userHandler.GetMe)
		auth.PATCH("/users/me", userHandler.UpdateMe)
		auth.DELETE("/users/me", userHandler.DeleteAccount)
		auth.PATCH("/users/me/privacy", userHandler.UpdatePrivacy)
		auth.GET("/users/me/interests", userHandler.GetInterests)
		auth.PUT("/users/me/interests", userHandler.SetInterests)
		auth.GET("/users/:identifier", userHandler.GetByUsername)

		// Follow & Social
		auth.POST("/users/:identifier/follow", socialHandler.Follow)
		auth.DELETE("/users/:identifier/follow", socialHandler.Unfollow)
		auth.GET("/users/:identifier/followers", socialHandler.Followers)
		auth.GET("/users/:identifier/following", socialHandler.Following)
		auth.DELETE("/users/:identifier/followers", socialHandler.RemoveFollower)
		auth.GET("/users/me/follow-requests", socialHandler.FollowRequests)
		auth.POST("/users/:identifier/follow-requests/approve", socialHandler.ApproveFollowRequest)
		auth.POST("/users/:identifier/follow-requests/decline", socialHandler.DeclineFollowRequest)

		// Close Friends [ADV]
		auth.POST("/users/:identifier/close-friends", socialHandler.AddCloseFriend)
		auth.DELETE("/users/:identifier/close-friends", socialHandler.RemoveCloseFriend)
		auth.GET("/users/me/close-friends", socialHandler.ListCloseFriends)

		// Block & Report
		auth.POST("/users/:identifier/block", socialHandler.Block)
		auth.DELETE("/users/:identifier/block", socialHandler.Unblock)
		auth.GET("/users/me/blocked", socialHandler.GetBlockedUsers)
		auth.POST("/reports", socialHandler.Report)

		// Posts
		auth.GET("/posts/:id", postHandler.GetByID)
		auth.POST("/posts", postHandler.Create)
		auth.PATCH("/posts/:id", postHandler.EditCaption)
		auth.GET("/feed", postHandler.Feed)
		auth.GET("/users/me/posts", postHandler.MyPosts)
		auth.GET("/users/:identifier/posts", postHandler.UserPosts)
		auth.DELETE("/posts/:id", postHandler.Delete)
		auth.POST("/posts/:id/archive", postHandler.Archive)
		auth.POST("/posts/:id/unarchive", postHandler.Unarchive)
		auth.DELETE("/posts/:id/tags/:taggedUserId", postHandler.RemoveTag)

		// Interactions (Like, Comment, Save, Share)
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

		// Stories
		auth.POST("/stories", storyHandler.Create)
		auth.DELETE("/stories/:id", storyHandler.Delete)
		auth.GET("/stories/feed", storyHandler.Feed)
		auth.GET("/stories/:id", storyHandler.GetByID)
		auth.POST("/stories/:id/views", storyHandler.RecordView)
		auth.GET("/stories/:id/viewers", storyHandler.Viewers)

		// Story Highlights [ADV]
		auth.GET("/highlights", highlightHandler.List)
		auth.POST("/highlights", highlightHandler.Create)
		auth.PATCH("/highlights/:id", highlightHandler.Update)
		auth.DELETE("/highlights/:id", highlightHandler.Delete)
		auth.POST("/highlights/:id/stories", highlightHandler.AddStory)
		auth.DELETE("/highlights/:id/stories/:story_id", highlightHandler.RemoveStory)

		// Search
		auth.GET("/search", searchHandler.Search)
		auth.GET("/hashtags/:tag/posts", searchHandler.HashtagPosts)

		// Search History [ADV]
		auth.GET("/search/history", searchHistoryHandler.List)
		auth.POST("/search/history", searchHistoryHandler.Save)
		auth.DELETE("/search/history/:id", searchHistoryHandler.DeleteItem)
		auth.DELETE("/search/history", searchHistoryHandler.DeleteAll)

		// Direct Messages
		auth.GET("/conversations", dmHandler.ListConversations)
		auth.GET("/conversations/requests", dmHandler.ListMessageRequests)
		auth.POST("/conversations", dmHandler.StartConversation)
		auth.GET("/conversations/:id/messages", dmHandler.ListMessages)
		auth.POST("/conversations/:id/messages", dmHandler.SendMessage)

		// Notifications
		auth.GET("/notifications", notifHandler.List)
		auth.POST("/notifications/read-all", notifHandler.ReadAll)
		auth.POST("/notifications", notifHandler.Create)
		auth.POST("/notifications/:id/read", notifHandler.Read)
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
