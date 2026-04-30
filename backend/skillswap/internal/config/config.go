package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl              string
	Port               string
	JWTSecret          string
	UploadDir          string
	BaseURL            string
	FrontendURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	ResendAPIKey       string
	FromEmail          string
	LiveKitAPIKey      string
	LiveKitAPISecret   string
	LiveKitURL         string

	// RequireEmailVerification, when true, gates write actions (creating swap
	// requests, etc.) on the requester having verified their email. Defaults
	// to false so existing unverified accounts keep working until ops flips it
	// after a bulk verify-now mailing. Set via REQUIRE_EMAIL_VERIFICATION=true.
	RequireEmailVerification bool
}

func Load() Config {
	_ = godotenv.Load()

	// Try DATABASE_URL first (Heroku format), then fall back to DB_URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
	}

	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	uploadDir := os.Getenv("UPLOAD_DIR")
	baseURL := os.Getenv("BASE_URL")
	frontendURL := os.Getenv("FRONTEND_URL")

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	resendAPIKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("FROM_EMAIL")

	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitAPISecret := os.Getenv("LIVEKIT_API_SECRET")
	livekitURL := os.Getenv("LIVEKIT_URL")

	requireEmailVerification := os.Getenv("REQUIRE_EMAIL_VERIFICATION") == "true"

	if dbURL == "" {
		log.Fatal("DATABASE_URL or DB_URL environment variable is required")
	}

	if jwtSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is required. Cannot start without a secure secret.")
	}

	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	if googleRedirectURL == "" && googleClientID != "" {
		googleRedirectURL = baseURL + "/api/v1/auth/google/callback"
	}

	if fromEmail == "" {
		fromEmail = "noreply@skillswap.com"
	}

	if port == "" {
		port = "8080"
	}

	return Config{
		DBUrl:              dbURL,
		Port:               port,
		JWTSecret:          jwtSecret,
		UploadDir:          uploadDir,
		BaseURL:            baseURL,
		FrontendURL:        frontendURL,
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		GoogleRedirectURL:  googleRedirectURL,
		ResendAPIKey:       resendAPIKey,
		FromEmail:          fromEmail,
		LiveKitAPIKey:      livekitAPIKey,
		LiveKitAPISecret:   livekitAPISecret,
		LiveKitURL:         livekitURL,

		RequireEmailVerification: requireEmailVerification,
	}
}
