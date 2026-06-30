// Package db provides a zerolog-backed GORM logger that bridges GORM's logging
// interface to the application's zerolog instance. This replaces the previous
// logger.Discard, enabling GORM error and slow-query logging through zerolog.
package db

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zercle/zercle-go-template/internal/config"
)

const defaultSlowThreshold = 200 * time.Millisecond

// gormLogger implements gorm.io/gorm/logger.Interface, bridging GORM's
// logging to a zerolog.Logger. It respects the configured log level and
// reports slow queries via the slowThreshold setting.
type gormLogger struct {
	log                       *zerolog.Logger
	level                     logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// newGORMLogger creates a GORM logger backed by the application's zerolog.
// The level is derived from cfg.Log.Level via zerolog.ParseLevel; if log is
// nil, a nop logger is used defensively (never panics).
func newGORMLogger(log *zerolog.Logger, cfg *config.Config) *gormLogger {
	if log == nil {
		nop := zerolog.Nop()
		log = &nop
	}

	level := logger.Info // default: info and above (info, warn, error)
	if cfg != nil {
		switch cfg.Log.Level {
		case "panic", "fatal", "error":
			level = logger.Error
		case "warn":
			level = logger.Warn
		default: // info, debug, trace — all log at Info level in GORM terms
			level = logger.Info
		}
	}

	return &gormLogger{
		log:                       log,
		level:                     level,
		slowThreshold:             defaultSlowThreshold,
		ignoreRecordNotFoundError: true,
	}
}

// LogMode returns a shallow copy of the logger with the given log level.
func (g *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &gormLogger{
		log:                       g.log,
		level:                     level,
		slowThreshold:             g.slowThreshold,
		ignoreRecordNotFoundError: g.ignoreRecordNotFoundError,
	}
}

// Info logs info-level messages. Only when the logger level allows it.
func (g *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	if g.level <= logger.Info {
		g.log.Info().Msgf(msg, args...)
	}
}

// Warn logs warning-level messages. Only when the logger level allows it.
func (g *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if g.level <= logger.Warn {
		g.log.Warn().Msgf(msg, args...)
	}
}

// Error logs error-level messages. Only when the logger level allows it.
func (g *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	if g.level <= logger.Error {
		g.log.Error().Msgf(msg, args...)
	}
}

// Trace logs SQL execution details: slow queries, errors, and the SQL itself.
func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		g.log.Error().Str("module", "gorm").Err(err).Dur("elapsed", elapsed).Int64("rows", rows).Msg(sql)

	case errors.Is(err, gorm.ErrRecordNotFound):
		if !g.ignoreRecordNotFoundError {
			g.log.Warn().Str("module", "gorm").Int64("rows", rows).Msg(sql)
		}

	case g.slowThreshold > 0 && elapsed > g.slowThreshold:
		g.log.Warn().Str("module", "gorm").Dur("elapsed", elapsed).Int64("rows", rows).Msgf("slow query: %s (threshold: %v)", sql, g.slowThreshold)

	default:
		if g.level <= logger.Info {
			g.log.Debug().Str("module", "gorm").Dur("elapsed", elapsed).Int64("rows", rows).Msg(sql)
		}
	}
}
