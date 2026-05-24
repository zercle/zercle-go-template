package telemetry

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with structured logging capabilities.
type Logger struct {
	zerolog.Logger
}

// New creates a new Logger with the specified level and format.
// Supported formats: "json" (default), "text".
func New(level, format string) *Logger {
	zlLevel := parseLevel(level)

	var output io.Writer = os.Stdout
	if format == "text" {
		output = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = time.RFC3339
			w.Out = os.Stdout
		})
	}

	zl := zerolog.New(output).
		Level(zlLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{zl}
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
