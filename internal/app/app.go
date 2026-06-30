// Package app is the reusable composition root. It wires the DI container and
// constructs a runnable server.Application for tests, CLIs, and the main entry
// point.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	exampledi "github.com/zercle/zercle-go-template/internal/features/example/di"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/server"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Version metadata is populated by cmd/server/main.go via these package-level
// variables before Run is called.
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

// Build wires the DI container in dependency order and returns the
// orchestrated application along with the populated injector.
//
// The sequence is config → telemetry → database → valkey → shared servers →
// example feature. On error the partially-wired injector is returned; the
// caller is responsible for calling injector.Shutdown() to release any
// providers that were successfully constructed.
func Build(ctx context.Context, cfg *config.Config) (*server.Application, do.Injector, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}

	injector := do.New()

	do.ProvideValue(injector, cfg)

	if err := telemetry.Register(ctx, injector); err != nil {
		return nil, injector, fmt.Errorf("register telemetry: %w", err)
	}

	logger, err := do.Invoke[*zerolog.Logger](injector)
	if err != nil {
		return nil, injector, fmt.Errorf("resolve logger: %w", err)
	}
	logger.Info().
		Str("version", Version).
		Str("commit", CommitSHA).
		Str("build_time", BuildTime).
		Str("env", cfg.App.Environment).
		Msg("starting server")

	if err := db.Register(ctx, injector); err != nil {
		return nil, injector, fmt.Errorf("register database: %w", err)
	}

	if err := valkey.Register(ctx, injector); err != nil {
		return nil, injector, fmt.Errorf("register valkey: %w", err)
	}

	if err := server.Register(injector); err != nil {
		return nil, injector, fmt.Errorf("register shared servers: %w", err)
	}

	if err := exampledi.Register(injector); err != nil {
		return nil, injector, fmt.Errorf("register example feature: %w", err)
	}

	application := server.NewApplication(injector, cfg, logger)
	return application, injector, nil
}

// Run builds the application and runs it until the context is cancelled or a
// server error occurs. It is the simplest entry point for tests and the main
// binary.
func Run(ctx context.Context, cfg *config.Config) error {
	application, injector, err := Build(ctx, cfg)
	if err != nil {
		if injector != nil {
			_ = injector.Shutdown()
		}
		return err
	}

	logger := application.Logger()
	defer func() {
		// samber/do v2's Injector.Shutdown returns *do.ShutdownReport (which
		// implements error), not a bare error. A non-nil report is returned
		// even on success, so we gate on report.Succeed rather than nilness;
		// treating the report as an error directly would log a spurious
		// failure on every clean shutdown.
		report := injector.Shutdown()
		if report != nil && !report.Succeed {
			logger.Error().Err(report).Msg("injector shutdown error")
		}
	}()

	if err := application.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run application: %w", err)
	}
	return nil
}
