// Package db wires PostgreSQL infrastructure into the DI container.
package db

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/telemetry"
)

// Register provides *gorm.DB and registers the PostgreSQL readiness checker.
// The ctx drives the initial DB construction so startup cancellation and
// connect timeouts propagate.
func Register(ctx context.Context, c do.Injector) error {
	cfg, err := do.Invoke[*config.Config](c)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}

	log, err := do.Invoke[*zerolog.Logger](c)
	if err != nil {
		return fmt.Errorf("resolve logger: %w", err)
	}

	db, err := NewDB(ctx, cfg, log)
	if err != nil {
		return err
	}
	do.ProvideValue(c, db)
	do.ProvideValue(c, NewShutdownCloser(db))

	registry, err := do.Invoke[*telemetry.Registry](c)
	if err != nil {
		return fmt.Errorf("resolve health registry: %w", err)
	}
	registry.AddReadiness(gormChecker{db: db})

	return nil
}
