package logging

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a structured logger configured for the service.
func New(service, env string) *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler).With("service", service, "env", env)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
