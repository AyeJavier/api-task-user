// Package config loads and holds application configuration from environment variables.
package config

import (
	"os"
	"strconv"
)

// Config holds all application settings.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
}

// AppConfig holds HTTP server settings.
type AppConfig struct {
	Env  string
	Port string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	URL      string
}

// JWTConfig holds token signing settings.
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// SecurityConfig holds security-related settings.
type SecurityConfig struct {
	BcryptCost              int
	RateLimitRequests       int
	RateLimitWindowSeconds  int
}

// Load reads configuration from environment variables.
// Panics if required values are missing.
func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     mustGetEnv("DB_NAME"),
			User:     mustGetEnv("DB_USER"),
			Password: mustGetEnv("DB_PASSWORD"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			URL:      mustGetEnv("DATABASE_URL"),
		},
		JWT: JWTConfig{
			Secret:          mustGetEnv("JWT_SECRET"),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		},
		Security: SecurityConfig{
			BcryptCost:             getEnvInt("BCRYPT_COST", 12),
			RateLimitRequests:      getEnvInt("RATE_LIMIT_REQUESTS", 5),
			RateLimitWindowSeconds: getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
