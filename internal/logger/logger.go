// Package logger provides structured logging capabilities for the application.
// It defines a Logger interface that abstracts the underlying logging implementation.
package logger

//go:generate mockgen -source=$GOFILE -destination=./mocks/$GOFILE -package=mocks

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// loggerKey is the context key for storing logger instances.
	loggerKey contextKey = "logger"
)

// Logger defines the interface for structured logging.
// Implementations should provide methods for different log levels and context propagation.
type Logger interface {
	// Debug logs a message at debug level.
	Debug(msg string, fields ...Field)
	// Info logs a message at info level.
	Info(msg string, fields ...Field)
	// Warn logs a message at warn level.
	Warn(msg string, fields ...Field)
	// Error logs a message at error level.
	Error(msg string, fields ...Field)
	// Fatal logs a message at fatal level and exits.
	Fatal(msg string, fields ...Field)

	// WithContext returns a logger with the given context.
	WithContext(ctx context.Context) Logger
	// WithFields returns a logger with additional fields.
	WithFields(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// String creates a string field.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field.
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field.
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field.
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any creates a field with any value.
func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// zerologLogger implements the Logger interface using zerolog.
type zerologLogger struct {
	logger zerolog.Logger
	ctx    context.Context
}

// New creates a new Logger instance with the given service name and environment.
func New(service, environment string) Logger {
	level := zerolog.InfoLevel
	if environment == "development" {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var zl zerolog.Logger
	if environment == "development" {
		zl = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	} else {
		zl = zerolog.New(os.Stdout)
	}

	zl = zl.With().
		Timestamp().
		Str("service", service).
		Str("environment", environment).
		Logger()

	return &zerologLogger{logger: zl}
}

// NewNop creates a no-op logger for testing.
func NewNop() Logger {
	return &zerologLogger{logger: zerolog.Nop()}
}

// Debug implements Logger.Debug.
func (l *zerologLogger) Debug(msg string, fields ...Field) {
	event := l.logger.Debug()
	addFields(event, fields...).Msg(msg)
}

// Info implements Logger.Info.
func (l *zerologLogger) Info(msg string, fields ...Field) {
	event := l.logger.Info()
	addFields(event, fields...).Msg(msg)
}

// Warn implements Logger.Warn.
func (l *zerologLogger) Warn(msg string, fields ...Field) {
	event := l.logger.Warn()
	addFields(event, fields...).Msg(msg)
}

// Error implements Logger.Error.
func (l *zerologLogger) Error(msg string, fields ...Field) {
	event := l.logger.Error()
	addFields(event, fields...).Msg(msg)
}

// Fatal implements Logger.Fatal.
func (l *zerologLogger) Fatal(msg string, fields ...Field) {
	event := l.logger.Fatal()
	addFields(event, fields...).Msg(msg)
}

// WithContext implements Logger.WithContext.
func (l *zerologLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}
	return &zerologLogger{
		logger: l.logger,
		ctx:    ctx,
	}
}

// WithFields implements Logger.WithFields.
func (l *zerologLogger) WithFields(fields ...Field) Logger {
	ctx := l.logger.With()
	for _, f := range fields {
		ctx = ctx.Interface(f.Key, f.Value)
	}
	return &zerologLogger{
		logger: ctx.Logger(),
		ctx:    l.ctx,
	}
}

// addFields adds fields to a zerolog event.
func addFields(event *zerolog.Event, fields ...Field) *zerolog.Event {
	for _, f := range fields {
		event = event.Interface(f.Key, f.Value)
	}
	return event
}

// FromContext extracts the logger from the context.
// Returns a no-op logger if no logger is found in context.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return NewNop()
	}
	if l, ok := ctx.Value(loggerKey).(Logger); ok {
		return l
	}
	return NewNop()
}

// WithContext adds the logger to the context.
func WithContext(ctx context.Context, l Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, loggerKey, l)
}
