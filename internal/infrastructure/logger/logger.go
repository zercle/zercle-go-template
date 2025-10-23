package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

type Logger struct {
	logger *zerolog.Logger
}

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

func (l *Logger) Info(msg string, fields ...any) {
	l.logger.Info().Fields(fields).Msg(msg)
}

func (l *Logger) Error(msg string, fields ...any) {
	l.logger.Error().Fields(fields).Msg(msg)
}

func (l *Logger) Debug(msg string, fields ...any) {
	l.logger.Debug().Fields(fields).Msg(msg)
}

func (l *Logger) Warn(msg string, fields ...any) {
	l.logger.Warn().Fields(fields).Msg(msg)
}

func (l *Logger) Fatal(msg string, fields ...any) {
	l.logger.Fatal().Fields(fields).Msg(msg)
}

func (l *Logger) With(key string, value any) *Logger {
	newLogger := l.logger.With().Str(key, fmt.Sprintf("%v", value)).Logger()
	return &Logger{logger: &newLogger}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.With("request_id", requestID)
}

func (l *Logger) WithUserID(userID string) *Logger {
	return l.With("user_id", userID)
}

func Fields(kv ...any) map[string]any {
	if len(kv)%2 != 0 {
		log.Warn().Msg("Fields: odd number of key-value pairs")
		return nil
	}

	result := make(map[string]any)
	for i := 0; i < len(kv); i += 2 {
		key := fmt.Sprintf("%v", kv[i])
		var value any
		if i+1 < len(kv) {
			value = kv[i+1]
		}
		result[key] = value
	}
	return result
}
