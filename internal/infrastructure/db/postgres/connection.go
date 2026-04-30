package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

// DB wraps a PostgreSQL connection pool with query helpers.
type DB struct {
	Pool    *pgxpool.Pool
	Queries *Queries
}

// NewConnection creates a new database connection pool.
func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MaxIdleConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Str("host", cfg.Host).Int("port", cfg.Port).Msg("Database connected")

	return &DB{Pool: pool, Queries: &Queries{db: pool}}, nil
}

// Close closes the database connection pool.
func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

// Ping checks the database connection is alive.
func (d *DB) Ping(ctx context.Context) error {
	return d.Pool.Ping(ctx)
}

// PoolStats returns connection pool statistics.
func (d *DB) PoolStats() *pgxpool.Stat {
	return d.Pool.Stat()
}
