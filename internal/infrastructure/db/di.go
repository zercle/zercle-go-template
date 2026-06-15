// Package db wires PostgreSQL infrastructure into the DI container.
package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	sqlcdb "github.com/zercle/zercle-go-template/internal/infrastructure/db/sqlc"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Register provides *pgxpool.Pool and *sqlcdb.Queries and registers the
// PostgreSQL readiness checker.
func Register(c do.Injector) error {
	do.Provide(c, func(i do.Injector) (*pgxpool.Pool, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return NewPool(context.Background(), cfg)
	})

	do.Provide(c, func(i do.Injector) (*sqlcdb.Queries, error) {
		pool, err := do.Invoke[*pgxpool.Pool](i)
		if err != nil {
			return nil, err
		}
		return sqlcdb.New(pool), nil
	})

	pool, err := do.Invoke[*pgxpool.Pool](c)
	if err != nil {
		return err
	}

	registry := do.MustInvoke[*telemetry.Registry](c)
	registry.AddReadiness(pgxChecker{pool: pool})

	return nil
}
