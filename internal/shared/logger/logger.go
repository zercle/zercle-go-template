package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	log zerolog.Logger
}

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

func (l *Logger) Debug() *zerolog.Event {
	return l.log.Debug()
}

func (l *Logger) Info() *zerolog.Event {
	return l.log.Info()
}

func (l *Logger) Warn() *zerolog.Event {
	return l.log.Warn()
}

func (l *Logger) Error() *zerolog.Event {
	return l.log.Error()
}

func (l *Logger) Fatal() *zerolog.Event {
	return l.log.Fatal()
}

func (l *Logger) With() zerolog.Context {
	return l.log.With()
}

func (l *Logger) Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

var defaultLogger *Logger

func Init(level, format string) error {
	logger, err := New(level, format)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

func Default() *Logger {
	if defaultLogger == nil {
		logger, _ := New("info", "json")
		return logger
	}
	return defaultLogger
}

func Debug() *zerolog.Event {
	return Default().Debug()
}

func Info() *zerolog.Event {
	return Default().Info()
}

func Warn() *zerolog.Event {
	return Default().Warn()
}

func Error() *zerolog.Event {
	return Default().Error()
}

func Fatal() *zerolog.Event {
	return Default().Fatal()
}

func With() zerolog.Context {
	return Default().With()
}

func Ctx(ctx context.Context) *zerolog.Logger {
	return Default().Ctx(ctx)
}
