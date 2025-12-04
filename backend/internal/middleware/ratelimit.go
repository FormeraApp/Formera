package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*clientLimit
	rate     int           // requests per window
	window   time.Duration // time window
	cleanup  time.Duration // cleanup interval for old entries
}

type clientLimit struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: maximum requests per window
// window: time window (e.g., 1 minute)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientLimit),
		rate:    rate,
		window:  window,
		cleanup: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from the given key should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	client, exists := rl.clients[key]
	if !exists || now.After(client.resetTime) {
		rl.clients[key] = &clientLimit{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if client.count >= rl.rate {
		return false
	}

	client.count++
	return true
}

// Remaining returns the number of remaining requests for a key
func (rl *RateLimiter) Remaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	client, exists := rl.clients[key]
	if !exists || time.Now().After(client.resetTime) {
		return rl.rate
	}

	remaining := rl.rate - client.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ResetTime returns when the rate limit resets for a key
func (rl *RateLimiter) ResetTime(key string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	client, exists := rl.clients[key]
	if !exists {
		return time.Now().Add(rl.window)
	}
	return client.resetTime
}

// cleanupLoop periodically removes expired entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, client := range rl.clients {
			if now.After(client.resetTime) {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitConfig holds configuration for the rate limit middleware
type RateLimitConfig struct {
	// Rate is the number of requests allowed per window
	Rate int
	// Window is the time window for rate limiting
	Window time.Duration
	// KeyFunc extracts the rate limit key from the request (default: IP address)
	KeyFunc func(*gin.Context) string
	// SkipFunc determines if rate limiting should be skipped for a request
	SkipFunc func(*gin.Context) bool
}

// DefaultKeyFunc returns the client IP as the rate limit key
func DefaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	if config.Rate <= 0 {
		config.Rate = 100 // default: 100 requests
	}
	if config.Window <= 0 {
		config.Window = time.Minute // default: per minute
	}
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}

	limiter := NewRateLimiter(config.Rate, config.Window)

	return func(c *gin.Context) {
		// Skip rate limiting if configured
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		key := config.KeyFunc(c)

		if !limiter.Allow(key) {
			remaining := limiter.Remaining(key)
			resetTime := limiter.ResetTime(key)

			c.Header("X-RateLimit-Limit", string(rune(config.Rate)))
			c.Header("X-RateLimit-Remaining", string(rune(remaining)))
			c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
			c.Header("Retry-After", resetTime.Sub(time.Now()).String())

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": resetTime.Sub(time.Now()).Seconds(),
			})
			return
		}

		// Add rate limit headers to response
		c.Header("X-RateLimit-Limit", string(rune(config.Rate)))
		c.Header("X-RateLimit-Remaining", string(rune(limiter.Remaining(key))))

		c.Next()
	}
}

// APIRateLimiter creates a rate limiter for general API endpoints
// Default: 100 requests per minute per IP
func APIRateLimiter() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		Rate:   100,
		Window: time.Minute,
	})
}

// AuthRateLimiter creates a stricter rate limiter for auth endpoints
// Default: 10 requests per minute per IP (to prevent brute force)
func AuthRateLimiter() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		Rate:   10,
		Window: time.Minute,
	})
}

// SubmissionRateLimiter creates a rate limiter for form submissions
// Default: 30 requests per minute per IP
func SubmissionRateLimiter() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		Rate:   30,
		Window: time.Minute,
	})
}
