// Package db wires PostgreSQL infrastructure into the DI container.
package db

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Register provides *gorm.DB and registers the PostgreSQL readiness checker.
// The ctx drives the initial DB construction so startup cancellation and
// connect timeouts propagate.
func Register(ctx context.Context, c do.Injector) error {
	cfg := do.MustInvoke[*config.Config](c)

	db, err := NewDB(ctx, cfg)
	if err != nil {
		return err
	}
	do.ProvideValue(c, db)

	registry, err := do.Invoke[*telemetry.Registry](c)
	if err != nil {
		return fmt.Errorf("resolve health registry: %w", err)
	}
	registry.AddReadiness(gormChecker{db: db})

	return nil
}
