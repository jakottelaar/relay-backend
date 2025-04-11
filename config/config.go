package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment       string
	Port              int
	DSN               string
	SupabaseJwtSecret string
	SupabaseUrl       string
	SupabaseApiKey    string
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

	cfg.SupabaseJwtSecret = getEnv("SUPABASE_JWT_SECRET", "")
	if cfg.SupabaseJwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET is required")
	}

	cfg.SupabaseUrl = getEnv("SUPABASE_URL", "")
	if cfg.SupabaseUrl == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}

	cfg.SupabaseApiKey = getEnv("SUPABASE_API_KEY", "")
	if cfg.SupabaseApiKey == "" {
		return nil, fmt.Errorf("SUPABASE_API_KEY is required")
	}

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
