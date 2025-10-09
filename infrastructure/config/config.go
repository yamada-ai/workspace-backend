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
func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default for local development
		dbURL = "postgres://localhost:5432/workspace?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	return &Config{
		DatabaseURL: dbURL,
		ServerPort:  fmt.Sprintf(":%s", port),
	}
}
