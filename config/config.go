package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret     string
	JWTExpiration int // hours

	// Firebase
	FirebaseCredentialsPath string

	// Encryption
	EncryptionKey string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		Environment:             getEnv("ENVIRONMENT", "development"),
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://localhost:5432/kwachatracker?sslmode=disable"),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:               getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiration:           getEnvInt("JWT_EXPIRATION_HOURS", 720), // 30 days
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS", "./firebase-credentials.json"),
		EncryptionKey:           getEnv("ENCRYPTION_KEY", "32-byte-key-for-aes-256-gcm!!!"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
