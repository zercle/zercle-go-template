// Package db wires PostgreSQL infrastructure into the DI container.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	sqlcdb "github.com/zercle/zercle-go-template/internal/infrastructure/db/sqlc"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Register provides *pgxpool.Pool and *sqlcdb.Queries and registers the
// PostgreSQL readiness checker. The ctx is used to drive the initial pool
// construction so startup cancellation/timeouts propagate.
func Register(ctx context.Context, c do.Injector) error {
	cfg := do.MustInvoke[*config.Config](c)

	pool, err := NewPool(ctx, cfg)
	if err != nil {
		return err
	}
	do.ProvideValue(c, pool)

	do.Provide(c, func(i do.Injector) (*sqlcdb.Queries, error) {
		pool, err := do.Invoke[*pgxpool.Pool](i)
		if err != nil {
			return nil, err
		}
		return sqlcdb.New(pool), nil
	})

	registry, err := do.Invoke[*telemetry.Registry](c)
	if err != nil {
		return fmt.Errorf("resolve health registry: %w", err)
	}
	registry.AddReadiness(pgxChecker{pool: pool})

	return nil
}
