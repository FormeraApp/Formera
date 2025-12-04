package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"formera/internal/config"
	"formera/internal/database"
	"formera/internal/handlers"
	"formera/internal/middleware"
	"formera/internal/pkg"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v\n\nPlease set a secure JWT_SECRET environment variable (at least 32 characters).", err)
	}

	// Initialize logger
	pkg.InitializeLogger(pkg.LoggerConfig{
		Level:  cfg.LogLevel,
		Pretty: cfg.LogPretty,
	})

	// Initialize database
	if err := database.Initialize(cfg.DBPath); err != nil {
		pkg.LogFatal().Err(err).Msg("Failed to initialize database")
	}

	// Initialize storage
	store, err := initStorage(cfg)
	if err != nil {
		pkg.LogFatal().Err(err).Msg("Failed to initialize storage")
	}
	pkg.LogInfo().Str("type", string(store.Type())).Msg("Storage initialized")

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
			pkg.LogFatal().Err(err).Msg("Invalid trusted proxies configuration")
		}
		pkg.LogInfo().Strs("proxies", cfg.TrustedProxies).Msg("Trusted proxies configured")
	}

	// Configure custom IP header if specified (e.g., CF-Connecting-IP for Cloudflare)
	if cfg.RealIPHeader != "" {
		r.RemoteIPHeaders = []string{cfg.RealIPHeader}
		pkg.LogInfo().Str("header", cfg.RealIPHeader).Msg("Using custom IP header")
	}

	r.Use(pkg.GinLogger())
	r.Use(pkg.GinRecovery())
	r.Use(middleware.SecurityHeaders())

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CorsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)
	formHandler := handlers.NewFormHandler()
	submissionHandler := handlers.NewSubmissionHandler()
	setupHandler := handlers.NewSetupHandler(cfg.JWTSecret)
	uploadHandler := handlers.NewUploadHandler(store, cfg.JWTSecret, cfg.ApiURL)
	userHandler := handlers.NewUserHandler()

	// Serve uploaded files - all files require handler (no direct static serving)
	// This ensures consistent behavior between local and S3 storage
	// Public access via /uploads/* for form backgrounds, logos, etc.
	// Protected access via /api/files/*?token=... for share links
	r.GET("/uploads/*path", uploadHandler.GetFilePublic)

	// Public routes with global rate limit (100 req/min per IP)
	api := r.Group("/api")
	api.Use(middleware.APIRateLimiter())
	{
		// Setup routes (public) - with strict rate limiting to prevent brute force
		api.GET("/setup/status", setupHandler.GetStatus)
		api.POST("/setup/complete", middleware.AuthRateLimiter(), setupHandler.CompleteSetup)

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

		// File serving endpoint with share token protection
		api.GET("/files/*path", uploadHandler.GetFileProtected)
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

		// File share URL generation (authenticated)
		protected.POST("/files/share", uploadHandler.GenerateShareURL)
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
		pkg.LogInfo().Str("port", cfg.Port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkg.LogFatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	pkg.LogInfo().Msg("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop cleanup scheduler
	cleanupScheduler.Stop()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		pkg.LogError().Err(err).Msg("Server forced to shutdown")
	}

	pkg.LogInfo().Msg("Server exited")
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

	pkg.LogInfo().Msg("Checking for local files to migrate to S3...")

	result, err := storage.MigrateLocalToS3(localPath, s3Store, cfg.Storage.DeleteAfterMigrate)
	if err != nil {
		pkg.LogError().Err(err).Msg("Migration error")
		return
	}

	if result.MigratedFiles > 0 {
		pkg.LogInfo().
			Int("files", result.MigratedFiles).
			Float64("size_mb", float64(result.MigratedBytes)/(1024*1024)).
			Msg("Migration complete")
	}

	if len(result.Errors) > 0 {
		pkg.LogWarn().Int("count", len(result.Errors)).Msg("Migration had errors")
		for _, e := range result.Errors {
			pkg.LogWarn().Str("error", e).Msg("Migration error detail")
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
