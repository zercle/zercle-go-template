package zerolog

import (
	"os"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// Init initializes the logger with the given level and format.
func Init(level, format string) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()

	return nil
}

// Error returns an error log event.
func Error() *zerolog.Event {
	return log.Error()
}

// Info returns an info log event.
func Info() *zerolog.Event {
	return log.Info()
}

// Debug returns a debug log event.
func Debug() *zerolog.Event {
	return log.Debug()
}

// Warn returns a warning log event.
func Warn() *zerolog.Event {
	return log.Warn()
}

// GetLogger returns the logger instance.
func GetLogger() zerolog.Logger {
	return log
}
