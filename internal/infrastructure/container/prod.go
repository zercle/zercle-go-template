package container

import (
	"database/sql"
	"fmt"

	// Import postgres driver for database/sql
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	postgresRepo "github.com/zercle/zercle-go-template/internal/adapter/storage/postgres"
	"github.com/zercle/zercle-go-template/internal/core/port"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// NewProdContainer initializes the full DI container for Production.
// It bundles the common container with Real Database Repositories.
func NewProdContainer(cfg *config.Config, log zerolog.Logger) (do.Injector, error) {
	injector := NewContainer(cfg, log)

	// Register Database Connection
	do.Provide(injector, func(i do.Injector) (*sql.DB, error) {
		cfg := do.MustInvoke[*config.Config](i)

		// Postgres DSN
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)

		database, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		database.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		database.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		database.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

		if err := database.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		return database, nil
	})

	// Register Real Repositories
	do.Provide(injector, func(i do.Injector) (port.UserRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgresRepo.NewUserRepository(db), nil
	})
	do.Provide(injector, func(i do.Injector) (port.PostRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgresRepo.NewPostRepository(db), nil
	})

	return injector, nil
}
