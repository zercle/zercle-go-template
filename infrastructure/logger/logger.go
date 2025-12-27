package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zercle/zercle-go-template/infrastructure/config"
)

// Logger wraps zerolog.Logger with additional context methods
type Logger struct {
	logger *zerolog.Logger
}

// NewLogger creates and configures a structured logger with the given logging configuration.
// It parses the log level, sets up the output format (console or JSON), and returns a configured Logger instance.
func NewLogger(cfg *config.LoggingConfig) *Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
			NoColor:    false,
		}
	} else {
		output = os.Stdout
	}

	logger := zerolog.New(output).With().Timestamp().Logger()

	return &Logger{
		logger: &logger,
	}
}

// Info logs an informational message with optional structured fields.
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.logger.Info().Fields(fields).Msg(msg)
}

// Error logs an error message with optional structured fields.
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.logger.Error().Fields(fields).Msg(msg)
}

// Debug logs a debug message with optional structured fields.
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug().Fields(fields).Msg(msg)
}

// Warn logs a warning message with optional structured fields.
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn().Fields(fields).Msg(msg)
}

// Fatal logs a fatal error message with optional structured fields and terminates the program.
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatal().Fields(fields).Msg(msg)
}

// With creates a new logger instance with an additional context key-value pair.
// The new logger inherits all previous context and adds the specified field.
func (l *Logger) With(key string, value interface{}) *Logger {
	newLogger := l.logger.With().Str(key, fmt.Sprintf("%v", value)).Logger()
	return &Logger{logger: &newLogger}
}

// WithRequestID creates a new logger instance with the request ID added to context.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.With("request_id", requestID)
}

// WithUserID creates a new logger instance with the user ID added to context.
func (l *Logger) WithUserID(userID string) *Logger {
	return l.With("user_id", userID)
}

// Fields converts alternating key-value pairs into a map for structured logging.
// Returns nil if an odd number of arguments is provided.
func Fields(kv ...interface{}) map[string]interface{} {
	if len(kv)%2 != 0 {
		log.Warn().Msg("Fields: odd number of key-value pairs")
		return nil
	}

	result := make(map[string]interface{})
	for i := 0; i < len(kv); i += 2 {
		key := fmt.Sprintf("%v", kv[i])
		var value interface{}
		if i+1 < len(kv) {
			value = kv[i+1]
		}
		result[key] = value
	}
	return result
}
