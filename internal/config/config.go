package config

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	Environment   string
	HTTPAddr      string
	LogLevel      slog.Level
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
}

func Load() (Config, error) {
	cfg := Config{
		Environment:   value("FUZE_ENV", "development"),
		HTTPAddr:      value("FUZE_HTTP_ADDR", ":8080"),
		DatabaseURL:   os.Getenv("FUZE_DATABASE_URL"),
		RedisAddr:     value("FUZE_REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("FUZE_REDIS_PASSWORD"),
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
