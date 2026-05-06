package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityConfig holds security headers configuration
type SecurityConfig struct {
	ContentTypeOptions      string
	FrameOptions            string
	XSSProtection           string
	ReferrerPolicy          string
	ContentSecurityPolicy   string
	StrictTransportSecurity string
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		ContentTypeOptions: "nosniff",
		FrameOptions:       "DENY",
		XSSProtection:      "1; mode=block",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		ContentSecurityPolicy:   "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline';",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
	}
}

// SecurityHeaders returns a configurable security headers middleware
func SecurityHeaders(config ...SecurityConfig) gin.HandlerFunc {
	cfg := DefaultSecurityConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		if cfg.ContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", cfg.ContentTypeOptions)
		}

		if cfg.FrameOptions != "" {
			c.Header("X-Frame-Options", cfg.FrameOptions)
		}

		if cfg.XSSProtection != "" {
			c.Header("X-XSS-Protection", cfg.XSSProtection)
		}

		if cfg.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", cfg.ReferrerPolicy)
		}

		if cfg.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}

		if cfg.StrictTransportSecurity != "" {
			c.Header("Strict-Transport-Security", cfg.StrictTransportSecurity)
		}

		c.Next()
	}
}
