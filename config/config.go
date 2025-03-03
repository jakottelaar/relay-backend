package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment         string
	Port                int
	DSN                 string
	JwtSecret           string
	JwtExpirationSecond int
}

func New() (*Config, error) {

	var cfg Config

	cfg.Environment = getEnv("ENVIRONMENT", "development")

	if cfg.Environment == "development" {
		err := godotenv.Load(".env.local")
		if err != nil {
			return nil, fmt.Errorf("Error loading .env.local file")
		}
	}

	cfg.Port = getEnvAsInt("PORT", 8080)

	cfg.DSN = getEnv("DSN", "")
	if cfg.DSN == "" {
		return nil, fmt.Errorf("DSN is required")
	}

	cfg.JwtSecret = getEnv("JWT_SECRET", "")
	if cfg.JwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	cfg.JwtExpirationSecond = getEnvAsInt("JWT_EXPIRATION_SECOND", 3600)

	return &cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
