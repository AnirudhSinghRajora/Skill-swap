package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig holds authentication middleware configuration
type AuthConfig struct {
	JWTSecret     string
	TokenLookup   string // "header:Authorization" or "query:token" or "cookie:jwt"
	TokenHeadName string // "Bearer"
	SkipPaths     []string
}

// DefaultAuthConfig returns default auth configuration
func DefaultAuthConfig(cfg config.Config) AuthConfig {
	return AuthConfig{
		JWTSecret:     cfg.JWTSecret,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		SkipPaths:     []string{},
	}
}

// JWTAuth returns a configurable JWT authentication middleware
func JWTAuth(cfg config.Config, authConfig ...AuthConfig) gin.HandlerFunc {
	config := DefaultAuthConfig(cfg)
	if len(authConfig) > 0 {
		config = authConfig[0]
	}

	return func(c *gin.Context) {
		// Skip authentication for specified paths
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		token, err := extractToken(c, config)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Parse and validate token
		jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil || !jwtToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims and set in context
		if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok && jwtToken.Valid {
			c.Set("user_id", claims["user_id"])
			c.Set("email", claims["email"])

			// Set admin flag if present
			if isAdmin, exists := claims["is_admin"]; exists {
				c.Set("is_admin", isAdmin)
			} else {
				c.Set("is_admin", false)
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminAuth middleware for admin-only routes
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		isAdminBool, ok := isAdmin.(bool)
		if !exists || !ok || !isAdminBool {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// OptionalAuth middleware that doesn't fail if no token is provided
func OptionalAuth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without setting user context
			c.Next()
			return
		}

		// Token provided, try to validate it
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// Invalid format, continue without setting user context
			c.Next()
			return
		}

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		// If valid, set user context
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user_id", claims["user_id"])
				c.Set("email", claims["email"])
				if isAdmin, exists := claims["is_admin"]; exists {
					c.Set("is_admin", isAdmin)
				}
			}
		}

		c.Next()
	}
}

// Helper function to extract token from request
func extractToken(c *gin.Context, config AuthConfig) (string, error) {
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid token lookup format")
	}

	switch parts[0] {
	case "header":
		authHeader := c.GetHeader(parts[1])
		if authHeader == "" {
			return "", fmt.Errorf("authorization header required")
		}

		tokenString := strings.TrimPrefix(authHeader, config.TokenHeadName+" ")
		if tokenString == authHeader {
			return "", fmt.Errorf("invalid authorization header format")
		}
		return tokenString, nil

	case "query":
		token := c.Query(parts[1])
		if token == "" {
			return "", fmt.Errorf("token query parameter required")
		}
		return token, nil

	case "cookie":
		cookie, err := c.Cookie(parts[1])
		if err != nil {
			return "", fmt.Errorf("token cookie required")
		}
		return cookie, nil

	default:
		return "", fmt.Errorf("unsupported token lookup method")
	}
}
