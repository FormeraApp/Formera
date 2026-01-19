package pkg

import (
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var Log zerolog.Logger

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string // debug, info, warn, error
	Pretty     bool   // Human-readable output (for development)
	TimeFormat string // Time format
}

// InitializeLogger sets up the global logger
func InitializeLogger(cfg LoggerConfig) {
	var output io.Writer = os.Stdout

	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	level := parseLevel(cfg.Level)

	Log = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// GinLogger returns a gin middleware for HTTP request logging
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		event := Log.Info()
		if status >= 400 && status < 500 {
			event = Log.Warn()
		} else if status >= 500 {
			event = Log.Error()
		}

		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP request")
	}
}

// GinRecovery returns a gin middleware for panic recovery with logging
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Log.Error().
					Interface("error", err).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Msg("Panic recovered")

				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

// Helper functions for common logging patterns

func LogDebug() *zerolog.Event {
	return Log.Debug()
}

func LogInfo() *zerolog.Event {
	return Log.Info()
}

func LogWarn() *zerolog.Event {
	return Log.Warn()
}

func LogError() *zerolog.Event {
	return Log.Error()
}

func LogFatal() *zerolog.Event {
	return Log.Fatal()
}

// WithRequestID adds request context to logs
func WithRequestID(requestID string) zerolog.Logger {
	return Log.With().Str("request_id", requestID).Logger()
}

// WithUserID adds user context to logs
func WithUserID(userID string) zerolog.Logger {
	return Log.With().Str("user_id", userID).Logger()
}
