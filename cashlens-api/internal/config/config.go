package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port            int
	Environment     string
	ShutdownTimeout time.Duration

	// Database
	DatabaseURL         string
	DBMaxConnections    int
	DBConnectionTimeout time.Duration

	// Clerk Auth
	ClerkPublishableKey string
	ClerkSecretKey      string

	// S3
	S3Bucket    string
	S3Region    string
	AWSEndpoint string // For LocalStack in development

	// Feature Flags
	EnableRateLimiting bool
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Port:                getEnvInt("PORT", 8080),
		Environment:         getEnv("ENVIRONMENT", "development"),
		ShutdownTimeout:     getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		DBMaxConnections:    getEnvInt("DB_MAX_CONNECTIONS", 25),
		DBConnectionTimeout: getEnvDuration("DB_CONNECTION_TIMEOUT", 30*time.Second),
		ClerkPublishableKey: getEnv("CLERK_PUBLISHABLE_KEY", ""),
		ClerkSecretKey:      getEnv("CLERK_SECRET_KEY", ""),
		S3Bucket:            getEnv("S3_BUCKET", ""),
		S3Region:            getEnv("S3_REGION", "ap-south-1"),
		AWSEndpoint:         getEnv("AWS_ENDPOINT", ""),
		EnableRateLimiting:  getEnvBool("ENABLE_RATE_LIMITING", false),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.ClerkSecretKey == "" && cfg.Environment == "production" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY is required in production")
	}
	if cfg.S3Bucket == "" && cfg.Environment == "production" {
		return nil, fmt.Errorf("S3_BUCKET is required in production")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
