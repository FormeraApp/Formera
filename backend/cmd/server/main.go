package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"formera/internal/config"
	"formera/internal/database"
	"formera/internal/handlers"
	"formera/internal/logger"
	"formera/internal/middleware"
	"formera/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// Swagger docs
	_ "formera/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Formera API
// @version 1.0
// @description REST API for Formera - a self-hosted form builder application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/your-repo/formera

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

func main() {
	cfg := config.Load()

	// Initialize logger
	logger.Initialize(logger.Config{
		Level:  cfg.LogLevel,
		Pretty: cfg.LogPretty,
	})

	// Initialize database
	if err := database.Initialize(cfg.DBPath); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}

	// Initialize storage
	store, err := initStorage(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	logger.Info().Str("type", string(store.Type())).Msg("Storage initialized")

	// Start cleanup scheduler
	cleanupScheduler := startCleanupScheduler(cfg, store)
	defer cleanupScheduler.Stop()

	// Setup Gin router with custom middleware
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Configure trusted proxies for accurate client IP detection
	// nil = trust all proxies (default), empty slice = trust no proxies
	if cfg.TrustedProxies == nil {
		r.SetTrustedProxies(nil) // Trust all (dev mode)
	} else if len(cfg.TrustedProxies) == 0 {
		r.SetTrustedProxies([]string{}) // Trust none
	} else {
		if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			logger.Fatal().Err(err).Msg("Invalid trusted proxies configuration")
		}
		logger.Info().Strs("proxies", cfg.TrustedProxies).Msg("Trusted proxies configured")
	}

	// Configure custom IP header if specified (e.g., CF-Connecting-IP for Cloudflare)
	if cfg.RealIPHeader != "" {
		r.RemoteIPHeaders = []string{cfg.RealIPHeader}
		logger.Info().Str("header", cfg.RealIPHeader).Msg("Using custom IP header")
	}

	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery())
	r.Use(middleware.SecurityHeaders())

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CorsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// Serve uploaded files - works for both local and S3 storage
	// For local storage: serves files directly from disk
	// For S3 storage: redirects to presigned URLs
	if cfg.Storage.GetStorageType() == "local" {
		r.Static("/uploads", cfg.Storage.LocalPath)
	} else {
		// For S3, use the upload handler to generate presigned URLs
		uploadHandlerForFiles := handlers.NewUploadHandler(store)
		r.GET("/uploads/*path", uploadHandlerForFiles.GetFile)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)
	formHandler := handlers.NewFormHandler()
	submissionHandler := handlers.NewSubmissionHandler()
	setupHandler := handlers.NewSetupHandler(cfg.JWTSecret)
	uploadHandler := handlers.NewUploadHandler(store)
	userHandler := handlers.NewUserHandler()

	// Public routes with global rate limit (100 req/min per IP)
	api := r.Group("/api")
	api.Use(middleware.APIRateLimiter())
	{
		// Setup routes (public)
		api.GET("/setup/status", setupHandler.GetStatus)
		api.POST("/setup/complete", setupHandler.CompleteSetup)

		// Auth routes with stricter rate limit (10 req/min per IP)
		api.POST("/auth/register", middleware.AuthRateLimiter(), authHandler.Register)
		api.POST("/auth/login", middleware.AuthRateLimiter(), authHandler.Login)

		// Public form access (supports both ID and slug)
		api.GET("/public/forms/:id", formHandler.GetPublic)
		api.POST("/public/forms/:id/verify-password", formHandler.VerifyPassword)

		// Form submission with moderate rate limit (30 req/min per IP)
		api.POST("/public/forms/:id/submit", middleware.SubmissionRateLimiter(), submissionHandler.Submit)

		// Public file upload (for form submissions with file fields)
		api.POST("/public/upload", uploadHandler.UploadFile)

		// File serving endpoint (redirects to S3 presigned URL or local file)
		api.GET("/files/*path", uploadHandler.GetFile)
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// User routes
		protected.GET("/auth/me", authHandler.Me)

		// Form routes
		protected.GET("/forms", formHandler.List)
		protected.POST("/forms", formHandler.Create)
		protected.GET("/forms/:id", formHandler.Get)
		protected.PUT("/forms/:id", formHandler.Update)
		protected.DELETE("/forms/:id", formHandler.Delete)
		protected.POST("/forms/:id/duplicate", formHandler.Duplicate)
		protected.GET("/forms/check-slug", formHandler.CheckSlugAvailability)

		// Submission routes
		protected.GET("/forms/:id/submissions", submissionHandler.List)
		protected.GET("/forms/:id/submissions/:submissionId", submissionHandler.Get)
		protected.DELETE("/forms/:id/submissions/:submissionId", submissionHandler.Delete)
		protected.GET("/forms/:id/stats", submissionHandler.Stats)
		protected.GET("/forms/:id/submissions/by-date", submissionHandler.SubmissionsByDate)
		protected.GET("/forms/:id/export/csv", submissionHandler.ExportCSV)
		protected.GET("/forms/:id/export/json", submissionHandler.ExportJSON)

		// Upload routes (authenticated)
		protected.POST("/uploads/image", uploadHandler.UploadImage)
		protected.POST("/uploads/file", uploadHandler.UploadFile)
		protected.DELETE("/uploads/:id", uploadHandler.DeleteFile)
	}

	// Admin routes (requires admin role)
	admin := api.Group("/")
	admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	admin.Use(middleware.AdminMiddleware())
	{
		// Settings routes (admin only)
		admin.GET("/settings", setupHandler.GetSettings)
		admin.PUT("/settings", setupHandler.UpdateSettings)

		// User management routes (admin only)
		admin.GET("/users", userHandler.List)
		admin.GET("/users/:id", userHandler.Get)
		admin.POST("/users", userHandler.Create)
		admin.PUT("/users/:id", userHandler.Update)
		admin.DELETE("/users/:id", userHandler.Delete)
	}

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/health/ready", func(c *gin.Context) {
		// Check database connection
		sqlDB, err := database.DB.DB()
		if err != nil {
			c.JSON(503, gin.H{"status": "error", "error": "database connection failed"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "error", "error": "database ping failed"})
			return
		}
		c.JSON(200, gin.H{"status": "ready", "database": "ok"})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("port", cfg.Port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop cleanup scheduler
	cleanupScheduler.Stop()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

// initStorage initializes the appropriate storage backend based on configuration
// and performs migration if needed
func initStorage(cfg *config.Config) (storage.Storage, error) {
	storageType := cfg.Storage.GetStorageType()

	switch storageType {
	case "s3":
		s3Store, err := storage.NewS3Storage(storage.S3Config{
			Bucket:          cfg.Storage.S3Bucket,
			Region:          cfg.Storage.S3Region,
			AccessKeyID:     cfg.Storage.S3AccessKeyID,
			SecretAccessKey: cfg.Storage.S3SecretAccessKey,
			Endpoint:        cfg.Storage.S3Endpoint,
			Prefix:          cfg.Storage.S3Prefix,
			PresignDuration: cfg.Storage.S3PresignDuration,
		})
		if err != nil {
			return nil, err
		}

		// Auto-migrate local files to S3 if enabled
		if cfg.Storage.MigrateOnStart {
			migrateLocalToS3(cfg, s3Store)
		}

		return s3Store, nil
	default:
		// ApiURL is the backend base URL (e.g., http://localhost:8080)
		uploadsURL := cfg.ApiURL + cfg.Storage.LocalURL
		return storage.NewLocalStorage(cfg.Storage.LocalPath, uploadsURL)
	}
}

// migrateLocalToS3 migrates existing local files to S3
func migrateLocalToS3(cfg *config.Config, s3Store *storage.S3Storage) {
	localPath := cfg.Storage.LocalPath

	// Check if local storage path exists and has files
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return // No local files to migrate
	}

	logger.Info().Msg("Checking for local files to migrate to S3...")

	result, err := storage.MigrateLocalToS3(localPath, s3Store, cfg.Storage.DeleteAfterMigrate)
	if err != nil {
		logger.Error().Err(err).Msg("Migration error")
		return
	}

	if result.MigratedFiles > 0 {
		logger.Info().
			Int("files", result.MigratedFiles).
			Float64("size_mb", float64(result.MigratedBytes)/(1024*1024)).
			Msg("Migration complete")
	}

	if len(result.Errors) > 0 {
		logger.Warn().Int("count", len(result.Errors)).Msg("Migration had errors")
		for _, e := range result.Errors {
			logger.Warn().Str("error", e).Msg("Migration error detail")
		}
	}
}

// startCleanupScheduler initializes and starts the file cleanup scheduler
func startCleanupScheduler(cfg *config.Config, store storage.Storage) *storage.CleanupScheduler {
	cleanupConfig := storage.CleanupConfig{
		Enabled:  cfg.Cleanup.Enabled,
		Interval: time.Duration(cfg.Cleanup.IntervalHours) * time.Hour,
		MinAge:   time.Duration(cfg.Cleanup.MinAgeDays) * 24 * time.Hour,
		DryRun:   cfg.Cleanup.DryRun,
	}

	scheduler := storage.NewCleanupScheduler(store, database.DB, cleanupConfig)
	scheduler.Start()

	return scheduler
}
