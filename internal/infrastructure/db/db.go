// Package db wires PostgreSQL infrastructure into the DI container.
package db

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zercle/zercle-go-template/internal/config"
)

// NewDB builds a configured *gorm.DB from the application config. It derives
// a DSN from cfg.DBConnString(), augments it with connect_timeout, opens the
// GORM connection, applies pool tuning via the underlying *sql.DB, and pings
// the database before returning. The caller is responsible for calling Close.
//
// Schema is owned by golang-migrate; AutoMigrate is never invoked here.
func NewDB(ctx context.Context, cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("build dsn: %w", err)
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Discard,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(int(cfg.DB.MaxConns))
	sqlDB.SetMaxIdleConns(int(cfg.DB.MinConns))
	sqlDB.SetConnMaxIdleTime(cfg.DB.MaxConnIdle)
	sqlDB.SetConnMaxLifetime(cfg.DB.MaxConnLife)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DB.ConnectTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return gormDB, nil
}

// buildDSN derives a DSN from cfg.DBConnString() and injects connect_timeout
// as an integer-second query parameter (minimum 1). pgx's stdlib driver honors
// connect_timeout, so the underlying transport respects the configured
// connect timeout without needing per-driver plumbing.
func buildDSN(cfg *config.Config) (string, error) {
	u, err := url.Parse(cfg.DBConnString())
	if err != nil {
		return "", fmt.Errorf("parse dsn: %w", err)
	}

	q := u.Query()
	seconds := int(cfg.DB.ConnectTimeout / time.Second)
	seconds = max(seconds, 1)
	q.Set("connect_timeout", strconv.Itoa(seconds))
	u.RawQuery = q.Encode()

	return u.String(), nil
}
