package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Max      int                       // Maximum number of requests
	Duration time.Duration             // Time window
	Message  string                    // Error message when rate limit exceeded
	KeyFunc  func(*gin.Context) string // Function to generate rate limit key
	Limiter  Limiter                   // Optional override; defaults to DefaultLimiter()
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:      100,         // 100 requests
		Duration: time.Minute, // per minute
		Message:  "Rate limit exceeded. Please try again later.",
		KeyFunc:  func(c *gin.Context) string { return c.ClientIP() },
	}
}

// RateLimit returns a rate limiting middleware backed by the configured Limiter
// (in-memory by default, pluggable via Limiter interface for distributed backends).
func RateLimit(config ...RateLimitConfig) gin.HandlerFunc {
	cfg := DefaultRateLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	limiter := cfg.Limiter
	if limiter == nil {
		limiter = DefaultLimiter()
	}

	return func(c *gin.Context) {
		key := cfg.KeyFunc(c)
		ok, resetAt := limiter.Allow(key, cfg.Max, cfg.Duration)
		if !ok {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       cfg.Message,
				"retry_after": int(time.Until(resetAt).Seconds()),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRateLimit returns a stricter rate limit for auth endpoints
func AuthRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Max:      5,           // 5 requests
		Duration: time.Minute, // per minute
		Message:  "Too many authentication attempts. Please try again later.",
		KeyFunc:  func(c *gin.Context) string { return c.ClientIP() },
	})
}
