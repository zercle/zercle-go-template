package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/sqlc/db"
)

type PostgresDatabase struct {
	Pool *pgxpool.Pool
}

func NewPostgresDatabase(cfg *config.DatabaseConfig) (*PostgresDatabase, error) {
	connString := buildPostgresConnectionString(cfg)

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

	return &PostgresDatabase{Pool: pool}, nil
}

func buildPostgresConnectionString(cfg *config.DatabaseConfig) string {
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

func (d *PostgresDatabase) Queries() interface{} {
	return db.New(d.Pool)
}

func (d *PostgresDatabase) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

func (d *PostgresDatabase) Ping(ctx context.Context) error {
	if d.Pool == nil {
		return errors.New("database pool is not initialized")
	}
	return d.Pool.Ping(ctx)
}

func (d *PostgresDatabase) Begin(ctx context.Context) (interface{}, error) {
	return d.Pool.Begin(ctx)
}
