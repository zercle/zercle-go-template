// Zerolog-backed GORM logger: it reports GORM errors and slow queries through
// the application's zerolog instance (see the package doc in db.go).
package db

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zercle/zercle-go-template/internal/platform/config"
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

	// GORM's LogLevel is ordered Silent < Error < Warn < Info: a message at
	// severity S is emitted when level >= S. Info is the most verbose setting,
	// Error the least besides Silent, matching gorm's own logger semantics.
	level := logger.Info
	if cfg != nil {
		switch cfg.Log.Level {
		case "panic", "fatal", "error":
			level = logger.Error
		case "warn":
			level = logger.Warn
		default: // info, debug, trace — log SQL statements too
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

// Info logs info-level messages when the configured level is at least Info.
func (g *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	if g.level >= logger.Info {
		g.log.Info().Msgf(msg, args...)
	}
}

// Warn logs warning-level messages when the configured level is at least Warn.
func (g *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if g.level >= logger.Warn {
		g.log.Warn().Msgf(msg, args...)
	}
}

// Error logs error-level messages when the configured level is at least Error.
func (g *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	if g.level >= logger.Error {
		g.log.Error().Msgf(msg, args...)
	}
}

// Trace logs SQL execution details: slow queries, errors, and the SQL itself.
func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		if g.level >= logger.Error {
			g.log.Error().Str("module", "gorm").Err(err).Dur("elapsed", elapsed).Int64("rows", rows).Msg(sql)
		}

	case errors.Is(err, gorm.ErrRecordNotFound):
		if !g.ignoreRecordNotFoundError && g.level >= logger.Warn {
			g.log.Warn().Str("module", "gorm").Int64("rows", rows).Msg(sql)
		}

	case g.slowThreshold > 0 && elapsed > g.slowThreshold:
		if g.level >= logger.Warn {
			g.log.Warn().Str("module", "gorm").Dur("elapsed", elapsed).Int64("rows", rows).Msgf("slow query: %s (threshold: %v)", sql, g.slowThreshold)
		}

	default:
		if g.level >= logger.Info {
			g.log.Debug().Str("module", "gorm").Dur("elapsed", elapsed).Int64("rows", rows).Msg(sql)
		}
	}
}
