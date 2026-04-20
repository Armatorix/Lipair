package config

import (
	"os"
	"strings"
)

type Config struct {
	ServerPort          string
	DatabaseURL         string
	JWTSecret           string
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	LichessClientID     string
	LichessClientSecret string
	LichessRedirectURL  string
	FrontendURL         string
	CORSOrigins         []string
}

func Load() *Config {
	corsOrigins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ",")
	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://chessmgr:chessmgr@localhost:5432/chessmgr?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "change-me-in-production"),
		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:   getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		LichessClientID:     getEnv("LICHESS_CLIENT_ID", ""),
		LichessClientSecret: getEnv("LICHESS_CLIENT_SECRET", ""),
		LichessRedirectURL:  getEnv("LICHESS_REDIRECT_URL", "http://localhost:8080/api/auth/lichess/callback"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		CORSOrigins:         corsOrigins,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
