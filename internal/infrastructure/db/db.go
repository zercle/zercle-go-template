// Package db wires PostgreSQL infrastructure into the DI container.
package db

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// NewDB builds a configured *gorm.DB from the application config. It opens the
// GORM connection using cfg.DBConnString(), applies pool tuning via the
// underlying *sql.DB, and pings the database before returning. The caller is
// responsible for closing the underlying *sql.DB obtained via (*gorm.DB).DB().
//
// Schema is owned by golang-migrate; AutoMigrate is never invoked here.
func NewDB(ctx context.Context, cfg *config.Config, log *zerolog.Logger) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	dsn := cfg.DBConnString()

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 newGORMLogger(log, cfg),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		// (*gorm.DB).DB() is the only accessor for the underlying *sql.DB; on
		// failure there is no handle to close, so the partially-opened pool is
		// abandoned. This path is unreachable in normal operation (it only
		// fails if the Dialector cannot provide a *sql.DB, which postgres
		// always can).
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(int(cfg.DB.MaxConns))
	sqlDB.SetMaxIdleConns(int(cfg.DB.MaxIdleConns))
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
