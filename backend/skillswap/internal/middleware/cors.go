package middleware

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	allowedOrigins := []string{"http://localhost:3000"}
	if origin := os.Getenv("FRONTEND_URL"); origin != "" {
		allowedOrigins = []string{origin}
	}
	return CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"Cache-Control",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// ConfigurableCORS returns a configurable CORS middleware
func ConfigurableCORS(config ...CORSConfig) gin.HandlerFunc {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders, ", "))

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
		}

		origin := c.GetHeader("Origin")
		if origin != "" {
			// Check if origin is allowed
			if contains(cfg.AllowOrigins, "*") || contains(cfg.AllowOrigins, origin) {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		} else if contains(cfg.AllowOrigins, "*") {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Helper functions
func joinStrings(slice []string, separator string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += separator + slice[i]
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
