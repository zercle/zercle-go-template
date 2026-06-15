// Package telemetry provides observability building blocks: zerolog logger,
// OpenTelemetry tracer, Prometheus metrics, and health probes.
package telemetry

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/zercle/zercle-go-template/internal/config"
)

// NewLogger builds a zerolog.Logger from configuration, sets the global level,
// and returns the configured logger. The logger writes JSON to stdout by default;
// switch to a human-readable console format when cfg.Log.Format is "console".
func NewLogger(cfg *config.Config) (*zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.Log.Level, err)
	}

	zerolog.SetGlobalLevel(level)

	var logger zerolog.Logger
	if cfg.Log.Format == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		logger = zerolog.New(os.Stdout)
	}

	logger = logger.With().Timestamp().Logger()

	return &logger, nil
}
