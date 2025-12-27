package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// Database wraps the connection pool with SQLC queries
type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase creates a new database connection pool
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	connString := buildConnectionString(cfg)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// buildConnectionString constructs a PostgreSQL connection string from the database configuration.
// The connection string uses the postgres:// scheme with SSL mode disabled.
func buildConnectionString(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Driver,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
}

// Queries returns a new SQLC queries instance bound to the database connection pool.
func (d *Database) Queries() *db.Queries {
	return db.New(d.Pool)
}

// Close gracefully closes the database connection pool and releases all resources.
func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

// Ping verifies that the database connection is alive and responsive.
func (d *Database) Ping(ctx context.Context) error {
	if d.Pool == nil {
		return errors.New("database pool is not initialized")
	}
	return d.Pool.Ping(ctx)
}
