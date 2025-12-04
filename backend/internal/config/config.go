package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Configuration errors
var (
	ErrJWTSecretRequired = errors.New("JWT_SECRET environment variable is required")
	ErrJWTSecretTooShort = errors.New("JWT_SECRET must be at least 32 characters")
	ErrJWTSecretInsecure = errors.New("JWT_SECRET is using an insecure default value")
)

func init() {
	// Try to load .env files in order of preference:
	// 1. Current directory (.env) - for production/Docker
	// 2. Parent directory (../.env) - for development when running from /backend
	envFiles := []string{".env", "../.env"}

	loaded := false
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			log.Printf("Loaded environment from %s", file)
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("No .env file found, using environment variables")
	}
}

type Config struct {
	Port       string
	BaseURL    string // Frontend URL (e.g., http://localhost:3000)
	ApiURL     string // Backend base URL (e.g., http://localhost:8080)
	DBPath     string
	JWTSecret  string
	CorsOrigin string

	// Logging configuration
	LogLevel  string // debug, info, warn, error
	LogPretty bool   // Human-readable output (for development)

	// Proxy configuration
	TrustedProxies  []string // List of trusted proxy IPs/CIDRs (empty = trust all)
	RealIPHeader    string   // Custom header for client IP (e.g., "CF-Connecting-IP", "X-Real-IP")

	// Storage configuration
	Storage StorageConfig

	// Cleanup configuration
	Cleanup CleanupConfig
}

type CleanupConfig struct {
	// Enabled determines if cleanup scheduler is active
	Enabled bool
	// IntervalHours between cleanup runs
	IntervalHours int
	// MinAgeDays is the minimum age of orphaned files before deletion
	MinAgeDays int
	// DryRun if true, only logs what would be deleted without actually deleting
	DryRun bool
}

type StorageConfig struct {
	// Type: "local" or "s3" (auto-detected if not set)
	Type string

	// Local storage settings
	LocalPath string
	LocalURL  string

	// S3 storage settings
	S3Bucket          string
	S3Region          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Endpoint        string        // Optional: for MinIO/S3-compatible services
	S3Prefix          string        // Optional: prefix for all files
	S3PresignDuration time.Duration // Optional: presigned URL duration

	// Migration settings
	MigrateOnStart     bool // Auto-migrate local files to S3 when S3 is enabled
	DeleteAfterMigrate bool // Delete local files after successful migration
}

// IsS3Configured returns true if S3 credentials are configured
func (s *StorageConfig) IsS3Configured() bool {
	return s.S3Bucket != "" && s.S3Region != "" && s.S3AccessKeyID != "" && s.S3SecretAccessKey != ""
}

// GetStorageType returns the effective storage type
func (s *StorageConfig) GetStorageType() string {
	if s.Type != "" {
		return s.Type
	}
	if s.IsS3Configured() {
		return "s3"
	}
	return "local"
}

// Load loads configuration from environment variables.
// Returns an error if JWT_SECRET is missing, too short, or insecure.
func Load() (*Config, error) {
	presignMinutes, _ := strconv.Atoi(getEnv("S3_PRESIGN_MINUTES", "60"))
	cleanupInterval, _ := strconv.Atoi(getEnv("CLEANUP_INTERVAL_HOURS", "24"))
	cleanupMinAge, _ := strconv.Atoi(getEnv("CLEANUP_MIN_AGE_DAYS", "7"))

	port := getEnv("PORT", "8080")
	baseURL := getEnv("BASE_URL", "http://localhost:3000")
	apiURL := getEnv("API_URL", "http://localhost:"+port)

	// Validate JWT_SECRET
	jwtSecret := getEnv("JWT_SECRET", "")
	if err := validateJWTSecret(jwtSecret); err != nil {
		return nil, err
	}

	// CORS_ORIGIN defaults to BASE_URL if not set (same-origin deployment)
	corsOrigin := getEnv("CORS_ORIGIN", "")
	if corsOrigin == "" {
		corsOrigin = baseURL
	}
	return &Config{
		Port:           port,
		BaseURL:        baseURL,
		ApiURL:         apiURL,
		DBPath:         getEnv("DB_PATH", "./data/formera.db"),
		JWTSecret:      jwtSecret,
		CorsOrigin:     corsOrigin,
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		LogPretty:      getEnv("LOG_PRETTY", "true") == "true",
		TrustedProxies: parseTrustedProxies(getEnv("TRUSTED_PROXIES", "")),
		RealIPHeader:   getEnv("REAL_IP_HEADER", ""),

		Storage: StorageConfig{
			Type: getEnv("STORAGE_TYPE", ""), // auto-detect if empty

			// Local storage
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./data/uploads"),
			LocalURL:  getEnv("STORAGE_LOCAL_URL", "/uploads"),

			// S3 storage
			S3Bucket:          getEnv("S3_BUCKET", ""),
			S3Region:          getEnv("S3_REGION", ""),
			S3AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
			S3SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
			S3Endpoint:        getEnv("S3_ENDPOINT", ""),
			S3Prefix:          getEnv("S3_PREFIX", ""),
			S3PresignDuration: time.Duration(presignMinutes) * time.Minute,

			// Migration
			MigrateOnStart:     getEnv("STORAGE_MIGRATE_ON_START", "true") == "true",
			DeleteAfterMigrate: getEnv("STORAGE_DELETE_AFTER_MIGRATE", "false") == "true",
		},

		Cleanup: CleanupConfig{
			Enabled:       getEnv("CLEANUP_ENABLED", "true") == "true",
			IntervalHours: cleanupInterval,
			MinAgeDays:    cleanupMinAge,
			DryRun:        getEnv("CLEANUP_DRY_RUN", "false") == "true",
		},
	}, nil
}

// validateJWTSecret checks if the JWT secret is secure
func validateJWTSecret(secret string) error {
	if secret == "" {
		return ErrJWTSecretRequired
	}
	if len(secret) < 32 {
		return ErrJWTSecretTooShort
	}
	// List of known insecure default values
	insecureDefaults := []string{
		"change-me-in-production-please",
		"secret",
		"your-secret-key",
		"changeme",
		"password",
	}
	secretLower := strings.ToLower(secret)
	for _, insecure := range insecureDefaults {
		if secretLower == insecure {
			return ErrJWTSecretInsecure
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// parseTrustedProxies parses a comma-separated list of trusted proxy IPs/CIDRs
func parseTrustedProxies(value string) []string {
	if value == "" {
		return nil // nil = trust all proxies (default for backwards compatibility)
	}
	var proxies []string
	for _, p := range strings.Split(value, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			proxies = append(proxies, p)
		}
	}
	return proxies
}
