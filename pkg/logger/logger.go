package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the global logger instance
var Log zerolog.Logger

// Config holds logger configuration
type Config struct {
	Level      string
	Pretty     bool
	TimeFormat string
}

// Setup initializes the global zerolog logger with configuration.
func Setup(cfg Config) {
	// Set global log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure time format (default: RFC3339)
	timeFormat := cfg.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}
	zerolog.TimeFieldFormat = timeFormat

	// Configure output writer
	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Initialize logger
	Log = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

// SetupDefault initializes logger with sensible defaults
func SetupDefault() {
	Setup(Config{
		Level:      "debug",
		Pretty:     false,
		TimeFormat: time.RFC3339,
	})
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.DebugLevel
	}
}

// Header is a helper type for structured logging context
type Header map[string]any

// NewHeader creates a new logging header
func NewHeader() Header {
	return make(Header)
}

// Add adds a key-value pair to the header
func (h Header) Add(key string, value any) Header {
	h[key] = value
	return h
}
