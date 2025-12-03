package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"formera/internal/config"
	"formera/internal/database"
	"formera/internal/handlers"
	"formera/internal/middleware"
	"formera/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize database
	if err := database.Initialize(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize storage
	store, err := initStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Printf("Storage initialized: %s", store.Type())

	// Start cleanup scheduler
	cleanupScheduler := startCleanupScheduler(cfg, store)
	defer cleanupScheduler.Stop()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cleanupScheduler.Stop()
		os.Exit(0)
	}()

	// Setup Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CorsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve static files for local storage
	if cfg.Storage.GetStorageType() == "local" {
		r.Static("/uploads", cfg.Storage.LocalPath)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)
	formHandler := handlers.NewFormHandler()
	submissionHandler := handlers.NewSubmissionHandler()
	setupHandler := handlers.NewSetupHandler(cfg.JWTSecret)
	uploadHandler := handlers.NewUploadHandler(store)
	userHandler := handlers.NewUserHandler()

	// Public routes
	api := r.Group("/api")
	{
		// Setup routes (public)
		api.GET("/setup/status", setupHandler.GetStatus)
		api.POST("/setup/complete", setupHandler.CompleteSetup)

		// Auth routes
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// Public form access (supports both ID and slug)
		api.GET("/public/forms/:id", formHandler.GetPublic)
		api.POST("/public/forms/:id/verify-password", formHandler.VerifyPassword)
		api.POST("/public/forms/:id/submit", submissionHandler.Submit)

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

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
		// Build full URL for local storage (BaseURL + LocalURL path)
		baseURL := cfg.BaseURL + cfg.Storage.LocalURL
		return storage.NewLocalStorage(cfg.Storage.LocalPath, baseURL)
	}
}

// migrateLocalToS3 migrates existing local files to S3
func migrateLocalToS3(cfg *config.Config, s3Store *storage.S3Storage) {
	localPath := cfg.Storage.LocalPath

	// Check if local storage path exists and has files
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return // No local files to migrate
	}

	log.Println("Checking for local files to migrate to S3...")

	result, err := storage.MigrateLocalToS3(localPath, s3Store, cfg.Storage.DeleteAfterMigrate)
	if err != nil {
		log.Printf("Migration error: %v", err)
		return
	}

	if result.MigratedFiles > 0 {
		log.Printf("Migration complete: %d files migrated (%.2f MB)",
			result.MigratedFiles, float64(result.MigratedBytes)/(1024*1024))
	}

	if len(result.Errors) > 0 {
		log.Printf("Migration had %d errors:", len(result.Errors))
		for _, e := range result.Errors {
			log.Printf("  - %s", e)
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
