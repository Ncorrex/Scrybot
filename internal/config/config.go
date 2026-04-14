package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Config holds the application configuration.
type Config struct {
	WebhookURL   string
	PollInterval time.Duration
	DataDir      string
	SearchQuery  string
}

// LoadFromEnv reads configuration from environment variables and applies defaults.
func LoadFromEnv() Config {
	cfg := Config{
		WebhookURL:   os.Getenv("WEBHOOK_URL"),
		PollInterval: parseDuration(os.Getenv("POLL_INTERVAL"), 1*time.Hour),
		DataDir:      os.Getenv("DATA_DIR"),
		SearchQuery:  os.Getenv("SEARCH_QUERY"),
	}

	if cfg.WebhookURL == "" {
		log.Fatal("WEBHOOK_URL environment variable is required")
	}

	if cfg.DataDir == "" {
		cfg.DataDir = "/root/data"
	}

	if cfg.SearchQuery == "" {
		cfg.SearchQuery = fmt.Sprintf("year=%d", time.Now().UTC().Year())
	}

	return cfg
}

// parseDuration parses a duration string with a fallback default.
func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("Warning: Invalid duration %q, using default %v", s, fallback)
		return fallback
	}
	return d
}
