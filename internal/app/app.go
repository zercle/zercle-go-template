// Package app is the reusable composition root. It wires the DI container and
// constructs a runnable server.Application for tests, CLIs, and the main entry
// point.
package app

import (
	"context"
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
// example feature. If any registration fails the injector is shut down and
// the error is returned with context.
func Build(ctx context.Context, cfg *config.Config) (*server.Application, do.Injector, error) {
	injector := do.New()

	do.ProvideValue(injector, cfg)

	if err := telemetry.Register(injector); err != nil {
		_ = injector.Shutdown()
		return nil, injector, fmt.Errorf("register telemetry: %w", err)
	}

	logger := do.MustInvoke[*zerolog.Logger](injector)
	logger.Info().
		Str("version", Version).
		Str("commit", CommitSHA).
		Str("build_time", BuildTime).
		Str("env", cfg.App.Environment).
		Msg("starting server")

	if err := db.Register(ctx, injector); err != nil {
		_ = injector.Shutdown()
		return nil, injector, fmt.Errorf("register database: %w", err)
	}

	if err := valkey.Register(ctx, injector); err != nil {
		_ = injector.Shutdown()
		return nil, injector, fmt.Errorf("register valkey: %w", err)
	}

	if err := server.Register(injector); err != nil {
		_ = injector.Shutdown()
		return nil, injector, fmt.Errorf("register shared servers: %w", err)
	}

	if err := exampledi.Register(injector); err != nil {
		_ = injector.Shutdown()
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
		return err
	}

	logger := application.Logger()
	defer func() {
		report := injector.Shutdown()
		if report != nil && !report.Succeed {
			logger.Error().Err(report).Msg("injector shutdown error")
		}
	}()

	if err := application.Run(ctx); err != nil {
		return fmt.Errorf("run application: %w", err)
	}
	return nil
}
