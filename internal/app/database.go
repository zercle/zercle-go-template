package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/infrastructure/database"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logging"
	"github.com/zercle/zercle-go-template/pkg/config"
)

// RegisterConfig registers the application configuration provider.
func RegisterConfig(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*config.Config, error) {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	})
}

// RegisterLogger registers the logger provider.
func RegisterLogger(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*logging.Logger, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := logging.New(cfg.Log)
		return logger, nil
	})
}

// RegisterDatabase registers the database connection pool provider.
func RegisterDatabase(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*database.DB, error) {
		cfg := do.MustInvoke[*config.Config](i)

		db, err := database.New(cfg.Database)
		if err != nil {
			return nil, err
		}

		return db, nil
	})

	// Also provide the raw pgxpool for repositories that need it
	do.Provide(i, func(i do.Injector) (*pgxpool.Pool, error) {
		db := do.MustInvoke[*database.DB](i)
		return db.Pool, nil
	})
}
