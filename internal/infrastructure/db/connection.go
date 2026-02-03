// Package db provides database connection management for the application.
// It handles PostgreSQL connection pooling and lifecycle management.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/infrastructure/db/sqlc"
	"zercle-go-template/internal/logger"
)

// DB wraps a PostgreSQL connection pool and provides sqlc querier access.
type DB struct {
	pool    *pgxpool.Pool
	querier sqlc.Querier
	logger  logger.Logger
}

// New creates a new database connection pool and querier.
// Returns nil without error if database configuration is not provided (for in-memory mode).
func New(cfg *config.Config, log logger.Logger) (*DB, error) {
	if cfg == nil || cfg.Database.Host == "" {
		log.Info("database configuration not provided, skipping database connection")
		return nil, nil
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute * 5

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := sqlc.New(pool)

	db := &DB{
		pool:    pool,
		querier: queries,
		logger:  log,
	}

	log.Info("database connection established successfully",
		logger.String("host", cfg.Database.Host),
		logger.Int("port", cfg.Database.Port),
		logger.String("database", cfg.Database.Database),
	)

	return db, nil
}

// Querier returns the sqlc querier for database operations.
func (d *DB) Querier() sqlc.Querier {
	return d.querier
}

// Close gracefully closes the database connection pool.
func (d *DB) Close() error {
	if d.pool != nil {
		d.logger.Info("closing database connection pool")
		d.pool.Close()
	}
	return nil
}

// Ping verifies the database connection is still alive.
func (d *DB) Ping(ctx context.Context) error {
	if d.pool == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return d.pool.Ping(ctx)
}

// Stats returns current connection pool statistics.
func (d *DB) Stats() *pgxpool.Stat {
	if d.pool == nil {
		return nil
	}
	return d.pool.Stat()
}
