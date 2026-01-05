package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kwachatracker/backend/config"
	"github.com/kwachatracker/backend/internal/database"
	"github.com/kwachatracker/backend/internal/handlers"
	"github.com/kwachatracker/backend/internal/middleware"
	"github.com/kwachatracker/backend/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Initialize Firebase Cloud Messaging (optional - fails gracefully)
	var fcmService *services.FCMService
	if _, err := os.Stat(cfg.FirebaseCredentialsPath); err == nil {
		fcmService, err = services.NewFCMService(cfg.FirebaseCredentialsPath)
		if err != nil {
			log.Printf("⚠️ FCM initialization failed (notifications disabled): %v", err)
		}
	} else {
		log.Println("⚠️ Firebase credentials not found, push notifications disabled")
	}

	// Initialize Gemini AI Service (optional - fails gracefully)
	var geminiService *services.GeminiService
	geminiService, err := services.NewGeminiService()
	if err != nil {
		log.Printf("⚠️ Gemini AI initialization failed (AI insights disabled): %v", err)
	} else {
		log.Println("✅ Gemini AI service initialized")
	}

	// Initialize handlers
	authHandler := &handlers.AuthHandler{Config: cfg}
	syncHandler := &handlers.SyncHandler{}
	analyticsHandler := &handlers.AnalyticsHandler{}

	// Initialize insights handler if Gemini is available
	var insightsHandler *handlers.InsightsHandler
	if geminiService != nil {
		insightsHandler = handlers.NewInsightsHandler(geminiService, fcmService)

		// Start daily analysis scheduler
		go startDailyScheduler(insightsHandler)
	}

	// Create router
	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimiter(100)) // 100 requests per minute

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"version":      "1.1.0",
			"time":         time.Now().Format(time.RFC3339),
			"ai_enabled":   geminiService != nil,
			"push_enabled": fcmService != nil,
		})
	})

	// Public routes
	r.POST("/api/v1/register", authHandler.Register)

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// User management
		protected.PUT("/consent", authHandler.UpdateConsent)
		protected.DELETE("/data", authHandler.DeleteData)

		// Transaction sync
		protected.POST("/sync", syncHandler.Sync)
		protected.GET("/transactions", syncHandler.GetTransactions)

		// Analytics
		protected.GET("/analytics/summary", analyticsHandler.GetSummary)
		protected.GET("/analytics/trends", analyticsHandler.GetTrends)

		// AI Insights (if Gemini is available)
		if insightsHandler != nil {
			protected.POST("/insights/generate", insightsHandler.GenerateInsights)
			protected.GET("/insights", insightsHandler.GetUserInsights)
		}

		// Push notifications (admin only)
		if fcmService != nil {
			protected.POST("/notify", func(c *gin.Context) {
				var req struct {
					Token string `json:"token" binding:"required"`
					Title string `json:"title" binding:"required"`
					Body  string `json:"body" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				err := fcmService.SendNotification(context.Background(), req.Token, req.Title, req.Body, nil)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
			})
		}
	}

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// startDailyScheduler runs AI analysis daily at 6 AM
func startDailyScheduler(handler *handlers.InsightsHandler) {
	log.Println("📅 Daily AI analysis scheduler started")

	for {
		// Calculate time until next 6 AM
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		duration := time.Until(next)

		log.Printf("⏰ Next AI analysis scheduled in %v", duration.Round(time.Minute))

		// Wait until scheduled time
		time.Sleep(duration)

		// Run daily analysis
		log.Println("🤖 Running scheduled AI analysis...")
		handler.RunDailyAnalysis()
	}
}
