package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/marcelofabianov/redtogreen/internal/platform/config"
)

func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	traceID, ok := ctx.Value("trace_id").(string)
	if ok && traceID != "" {
		return logger.With("trace_id", traceID)
	}
	return logger
}

func NewSlogLogger(cfg config.LoggerConfig) *slog.Logger {
	var level slog.Level

	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler)
}
