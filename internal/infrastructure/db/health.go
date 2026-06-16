package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxChecker reports PostgreSQL connectivity via pool.Ping.
type pgxChecker struct {
	pool *pgxpool.Pool
}

// Name returns the dependency name reported in health output.
func (pgxChecker) Name() string {
	return "postgres"
}

// Check verifies PostgreSQL is reachable by pinging the connection pool.
func (c pgxChecker) Check(ctx context.Context) error {
	if err := c.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}
