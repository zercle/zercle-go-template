// Package db configures and connects to PostgreSQL via pgx.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zercle/zercle-go-template/internal/config"
)

// NewPool builds a configured *pgxpool.Pool from the application config.
// It parses the DSN, applies pool tuning parameters, creates the pool, and
// pings it before returning. The caller is responsible for calling Close.
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DBConnString())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	poolCfg.MaxConns = cfg.DB.MaxConns
	poolCfg.MinConns = cfg.DB.MinConns
	poolCfg.MaxConnIdleTime = cfg.DB.MaxConnIdle
	poolCfg.MaxConnLifetime = cfg.DB.MaxConnLife
	poolCfg.ConnConfig.ConnectTimeout = cfg.DB.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DB.ConnectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
