package di

import (
	"github.com/samber/do"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

// ProvideConfig loads and provides the application configuration.
func ProvideConfig(i *do.Injector) (*config.Config, error) {
	cfg, err := config.Load("./configs")
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// ProvideLogger initializes and provides the logger using the loaded configuration.
func ProvideLogger(i *do.Injector) (struct{}, error) {
	cfg := do.MustInvoke[*config.Config](i)
	_ = logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	return struct{}{}, nil
}

// ProvideDatabase creates and provides a PostgreSQL database connection.
func ProvideDatabase(i *do.Injector) (*postgres.DB, error) {
	cfg := do.MustInvoke[*config.Config](i)
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// ProvideValkey creates and provides a Valkey client connection.
func ProvideValkey(i *do.Injector) (*valkey.Client, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return valkey.New(cfg.Valkey)
}

// ProvidePubSubClient creates and provides a Valkey PubSub client.
func ProvidePubSubClient(i *do.Injector) (valkey.PubSubClient, error) {
	return ProvideValkey(i)
}

// RegisterRootProviders registers all root-level dependency providers into the container.
func RegisterRootProviders(i *do.Injector) {
	do.Provide(i, ProvideConfig)
	do.Provide(i, ProvideLogger)
	do.Provide(i, ProvideDatabase)
	do.Provide(i, ProvideValkey)
	do.Provide(i, ProvidePubSubClient)
}
