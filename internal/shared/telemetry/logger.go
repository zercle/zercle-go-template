package telemetry

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with structured logging capabilities.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger with the specified level and format.
func New(level, format string) (*Logger, error) {
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &Logger{slog.New(handler)}, nil
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
