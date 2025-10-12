package config

import (
	"fmt"
	"os"
)

// Config holds application configuration
type Config struct {
	DatabaseURL string
	ServerPort  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	return &Config{
		DatabaseURL: dbURL,
		ServerPort:  fmt.Sprintf(":%s", port),
	}, nil
}
