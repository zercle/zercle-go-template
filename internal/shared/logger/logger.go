package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog to provide structured logging with context support.
type Logger struct {
	log zerolog.Logger
}

// New creates a new Logger with the specified log level and format.
func New(level, format string) (*Logger, error) {
	var output io.Writer = os.Stdout

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339
	logger := zerolog.New(output).Level(lvl)

	if format == "json" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return &Logger{log: logger}, nil
}

// Debug returns a debug level event for structured logging.
func (l *Logger) Debug() *zerolog.Event {
	return l.log.Debug()
}

// Info returns an info level event for structured logging.
func (l *Logger) Info() *zerolog.Event {
	return l.log.Info()
}

// Warn returns a warn level event for structured logging.
func (l *Logger) Warn() *zerolog.Event {
	return l.log.Warn()
}

func (l *Logger) Error() *zerolog.Event {
	return l.log.Error()
}

// Fatal returns a fatal level event for structured logging.
func (l *Logger) Fatal() *zerolog.Event {
	return l.log.Fatal()
}

// With returns a context builder for adding structured fields.
func (l *Logger) With() zerolog.Context {
	return l.log.With()
}

// Ctx returns a logger enriched with context for request tracing.
func (l *Logger) Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

var defaultLogger *Logger

// Init initializes the default global logger with the specified level and format.
func Init(level, format string) error {
	logger, err := New(level, format)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// Default returns the default global logger, initializing it with "info" and "json" if nil.
func Default() *Logger {
	if defaultLogger == nil {
		logger, _ := New("info", "json")
		return logger
	}
	return defaultLogger
}

// Debug logs a debug level event using the default global logger.
func Debug() *zerolog.Event {
	return Default().Debug()
}

// Info logs an info level event using the default global logger.
func Info() *zerolog.Event {
	return Default().Info()
}

// Warn logs a warn level event using the default global logger.
func Warn() *zerolog.Event {
	return Default().Warn()
}

// Error logs an error level event using the default global logger.
func Error() *zerolog.Event {
	return Default().Error()
}

// Fatal logs a fatal level event using the default global logger.
func Fatal() *zerolog.Event {
	return Default().Fatal()
}

// With creates a context builder using the default global logger.
func With() zerolog.Context {
	return Default().With()
}

// Ctx retrieves a context-enriched logger using the default global logger.
func Ctx(ctx context.Context) *zerolog.Logger {
	return Default().Ctx(ctx)
}
