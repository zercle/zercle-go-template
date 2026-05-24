package di

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// ProvideConfig creates and provides the application configuration.
func ProvideConfig(i do.Injector) (*config.Config, error) {
	cfg := config.Load()
	return &cfg, nil
}

// ProvideLogger creates and provides a structured logger.
func ProvideLogger(i do.Injector) (*telemetry.Logger, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return telemetry.New(cfg.LogLevel, cfg.LogFormat), nil
}

// ProvideDatabase creates and provides a PostgreSQL connection pool.
func ProvideDatabase(i do.Injector) (*postgres.DB, error) {
	cfg := do.MustInvoke[*config.Config](i)
	db, err := postgres.NewPool(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}
	return db, nil
}

// ProvideValkey creates and provides a Valkey client.
func ProvideValkey(i do.Injector) (*valkey.Client, error) {
	cfg := do.MustInvoke[*config.Config](i)
	logger := do.MustInvoke[*telemetry.Logger](i)
	client, err := valkey.New(valkey.ValkeyConfig{
		Host:     cfg.CacheHost,
		Port:     cfg.CachePort,
		Password: cfg.CachePassword,
		DB:       cfg.CacheDB,
	}, &logger.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}
	return client, nil
}

// ProvidePubSubClient creates and provides a PubSub client backed by Valkey.
func ProvidePubSubClient(i do.Injector) (valkey.PubSubClient, error) {
	return ProvideValkey(i)
}

// RegisterRootProviders registers all root-level DI providers.
func RegisterRootProviders(i do.Injector) {
	do.Provide(i, ProvideConfig)
	do.Provide(i, ProvideLogger)
	do.Provide(i, ProvideDatabase)
	do.Provide(i, ProvideValkey)
	do.Provide(i, ProvidePubSubClient)
}
