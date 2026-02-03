package di

import (
	"github.com/samber/do"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

func ProvideConfig(i *do.Injector) (*config.Config, error) {
	cfg, err := config.Load("./configs")
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func ProvideLogger(i *do.Injector) (struct{}, error) {
	cfg := do.MustInvoke[*config.Config](i)
	_ = logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	return struct{}{}, nil
}

func ProvideDatabase(i *do.Injector) (*postgres.DB, error) {
	cfg := do.MustInvoke[*config.Config](i)
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func ProvideValkey(i *do.Injector) (*valkey.Client, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return valkey.New(cfg.Valkey)
}

func ProvidePubSubClient(i *do.Injector) (valkey.PubSubClient, error) {
	return ProvideValkey(i)
}

func RegisterRootProviders(i *do.Injector) {
	do.Provide(i, ProvideConfig)
	do.Provide(i, ProvideLogger)
	do.Provide(i, ProvideDatabase)
	do.Provide(i, ProvideValkey)
	do.Provide(i, ProvidePubSubClient)
}
