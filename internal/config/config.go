package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Environment       string
	HTTPAddr          string
	LogLevel          slog.Level
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	CookieSecure      bool
	GitHubClientID    string
	GitHubClientSecret string
	GitHubCallbackURL string
}

func Load() (Config, error) {
	secure, err := strconv.ParseBool(value("FUZE_COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid FUZE_COOKIE_SECURE: %w", err)
	}

	cfg := Config{
		Environment:        value("FUZE_ENV", "development"),
		HTTPAddr:           value("FUZE_HTTP_ADDR", ":8080"),
		DatabaseURL:        os.Getenv("FUZE_DATABASE_URL"),
		RedisAddr:          value("FUZE_REDIS_ADDR", "localhost:6379"),
		RedisPassword:      os.Getenv("FUZE_REDIS_PASSWORD"),
		CookieSecure:       secure,
		GitHubClientID:     os.Getenv("FUZE_GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("FUZE_GITHUB_CLIENT_SECRET"),
		GitHubCallbackURL:  os.Getenv("FUZE_GITHUB_CALLBACK_URL"),
	}

	level, err := parseLevel(value("FUZE_LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level
	return cfg, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLevel(value string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return 0, fmt.Errorf("invalid log level %q: %w", value, err)
	}
	return level, nil
}
