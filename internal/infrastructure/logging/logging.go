// Package logging provides structured logging using zerolog.
package logging

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/zercle/zercle-go-template/pkg/config"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// requestIDKey is the context key for request ID.
	requestIDKey contextKey = "requestID"
	// userIDKey is the context key for user ID.
	userIDKey contextKey = "userID"
)

// Logger wraps zerolog.Logger for application-wide logging.
type Logger struct {
	zerolog.Logger
}

// New creates a new Logger with the given configuration.
func New(cfg config.LogConfig) *Logger {
	var output io.Writer = os.Stdout

	// Pretty printing for development (console output)
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global level based on config
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure error stack marshaling
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: logger}
}

// parseLevel converts a string level to zerolog.Level.
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithContext returns a logger with context values (request_id, user_id) attached.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.With()

	// Add request ID if present
	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		logger = logger.Str("request_id", requestID)
	}

	// Add user ID if present
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		logger = logger.Str("user_id", userID)
	}

	return &Logger{Logger: logger.Logger()}
}

// ToContext adds the logger to context.
func (l *Logger) ToContext(ctx context.Context) context.Context {
	// Use a package-level key to avoid collisions
	return context.WithValue(ctx, &loggerKey{}, l)
}

// loggerKey is the context key for the logger itself.
type loggerKey struct{}

// FromContext retrieves the logger from context or returns a new one.
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(&loggerKey{}).(*Logger); ok {
		return logger
	}
	return defaultLogger()
}

// defaultLogger returns a default logger for fallback use.
func defaultLogger() *Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	return &Logger{Logger: zerolog.Nop()}
}

// WithRequestID returns a new context with the request ID set.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID returns a new context with the user ID set.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// Debug logs a debug message with optional fields.
func (l *Logger) Debug() *zerolog.Event {
	return l.Logger.Debug()
}

// Info logs an info message with optional fields.
func (l *Logger) Info() *zerolog.Event {
	return l.Logger.Info()
}

// Warn logs a warning message with optional fields.
func (l *Logger) Warn() *zerolog.Event {
	return l.Logger.Warn()
}

// Error logs an error message with optional fields.
func (l *Logger) Error() *zerolog.Event {
	return l.Logger.Error()
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal() *zerolog.Event {
	return l.Logger.Fatal()
}

// Hook is an interface for zerolog event hooks.
type Hook interface {
	Run(e *zerolog.Event, level zerolog.Level, msg string)
}
