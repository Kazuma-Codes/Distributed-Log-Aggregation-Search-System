// Package config provides application configuration loaded from environment variables.
//
// All configuration is read at startup and validated. Sensible defaults are
// provided so the service can run locally without any env vars set.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration values sourced from environment
// variables. Zero-value fields carry the documented defaults.
type Config struct {
	// ClickHouse connection settings.
	ClickHouseAddr     string
	ClickHouseDatabase string
	ClickHouseUsername  string
	ClickHousePassword string

	// HTTP server settings.
	APIPort      int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Authentication. An empty APIKey disables authentication (dev mode).
	APIKey string

	// Rate limiting.
	RateLimitRPS   float64
	RateLimitBurst int

	// Logging.
	LogLevel string
}

// Load reads configuration from environment variables and returns a validated
// Config. It returns an error only if a value cannot be parsed.
func Load() (*Config, error) {
	cfg := &Config{
		ClickHouseAddr:     envOrDefault("CLICKHOUSE_ADDR", "localhost:9000"),
		ClickHouseDatabase: envOrDefault("CLICKHOUSE_DATABASE", "default"),
		ClickHouseUsername:  envOrDefault("CLICKHOUSE_USERNAME", "default"),
		ClickHousePassword: envOrDefault("CLICKHOUSE_PASSWORD", ""),
		APIKey:             envOrDefault("API_KEY", ""),
		LogLevel:           envOrDefault("LOG_LEVEL", "info"),
	}

	var err error

	cfg.APIPort, err = envOrDefaultInt("API_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid API_PORT: %w", err)
	}

	cfg.ReadTimeout, err = envOrDefaultDuration("READ_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}

	cfg.WriteTimeout, err = envOrDefaultDuration("WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}

	cfg.RateLimitRPS, err = envOrDefaultFloat("RATE_LIMIT_RPS", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
	}

	cfg.RateLimitBurst, err = envOrDefaultInt("RATE_LIMIT_BURST", 200)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BURST: %w", err)
	}

	return cfg, nil
}

// envOrDefault returns the value of the named environment variable or
// fallback if the variable is unset or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envOrDefaultInt parses an integer environment variable with a fallback.
func envOrDefaultInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return strconv.Atoi(v)
}

// envOrDefaultFloat parses a float64 environment variable with a fallback.
func envOrDefaultFloat(key string, fallback float64) (float64, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(v, 64)
}

// envOrDefaultDuration parses a duration environment variable (e.g. "30s")
// with a fallback.
func envOrDefaultDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return time.ParseDuration(v)
}
