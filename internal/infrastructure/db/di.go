// Package db wires PostgreSQL infrastructure into the DI container.
package db

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Register provides *gorm.DB and registers the PostgreSQL readiness checker.
// The ctx drives the initial DB construction so startup cancellation and
// connect timeouts propagate.
func Register(ctx context.Context, c do.Injector) error {
	cfg := do.MustInvoke[*config.Config](c)

	log, err := do.Invoke[*zerolog.Logger](c)
	if err != nil {
		return fmt.Errorf("resolve logger: %w", err)
	}

	db, err := NewDB(ctx, cfg, log)
	if err != nil {
		return err
	}
	do.ProvideValue(c, db)
	// NewShutdowner and the Shutdowner struct live in shutdowner.go (same
	// package); they adapt *gorm.DB to do's ShutdownerWithContextAndError so
	// injector.Shutdown() closes the connection pool.
	do.ProvideValue(c, NewShutdowner(db))

	registry, err := do.Invoke[*telemetry.Registry](c)
	if err != nil {
		return fmt.Errorf("resolve health registry: %w", err)
	}
	registry.AddReadiness(gormChecker{db: db})

	return nil
}
