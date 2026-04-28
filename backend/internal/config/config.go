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
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://lipair:lipair@localhost:5432/lipair?sslmode=disable"),
		JWTSecret:           getSecret("JWT_SECRET", "/run/secrets/jwt_secret", "change-me-in-production"),
		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  getSecret("GOOGLE_CLIENT_SECRET", "/run/secrets/google_client_secret", ""),
		GoogleRedirectURL:   getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		LichessClientID:     getEnv("LICHESS_CLIENT_ID", ""),
		LichessClientSecret: getSecret("LICHESS_CLIENT_SECRET", "/run/secrets/lichess_client_secret", ""),
		LichessRedirectURL:  getEnv("LICHESS_REDIRECT_URL", "http://localhost:8080/api/auth/lichess/callback"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		CORSOrigins:         corsOrigins,
	}
}

// getEnv returns the environment variable value or a default.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getSecret reads a secret value first from an environment variable, then from a
// Docker secret file (for Docker Swarm deployments), and finally falls back to the default.
func getSecret(envKey, secretFile, defaultValue string) string {
	if value, exists := os.LookupEnv(envKey); exists {
		return value
	}
	if data, err := os.ReadFile(secretFile); err == nil {
		return strings.TrimSpace(string(data))
	}
	return defaultValue
}
