package config

import "os"

// Config holds configuration for the desktop POS application.
type Config struct {
	APIBaseURL   string // Default API URL (overridden by stored config)
	DatabasePath string // Path to the local SQLite database file
	SyncInterval int    // Sync interval in seconds
}

// Load reads POS configuration from environment variables.
func Load() *Config {
	return &Config{
		APIBaseURL:   getEnv("API_BASE_URL", "http://localhost:8080"),
		DatabasePath: getEnv("DB_PATH", "coffeeshop.db"),
		SyncInterval: 30, // 30 seconds default
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
