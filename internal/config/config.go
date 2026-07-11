// Package config provides application configuration loading from environment variables.
package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	DatabaseURL string
	ServerPort  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	cfg := Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
	}

	if port, err := strconv.Atoi(cfg.ServerPort); err != nil || port < 1 || port > 65535 {
		log.Printf("invalid SERVER_PORT '%s', using default 8080", cfg.ServerPort)
		cfg.ServerPort = "8080"
	}

	return cfg
}

// getEnv returns the value of the environment variable or a fallback if not set.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
