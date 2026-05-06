package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl     string
	Port      string
	JWTSecret string
	UploadDir string
	BaseURL   string
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

	if port == "" {
		port = "8080"
	}

	return Config{
		DBUrl:     dbURL,
		Port:      port,
		JWTSecret: jwtSecret,
		UploadDir: uploadDir,
		BaseURL:   baseURL,
	}
}
