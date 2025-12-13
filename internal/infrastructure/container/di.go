package container

import (
	"database/sql"
	"fmt"

	// Import postgres driver for database/sql
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// NewContainer initializes the full DI container.
func NewContainer(cfg *config.Config, log zerolog.Logger) (do.Injector, error) {
	injector := do.New()

	// 1. Config
	do.ProvideValue(injector, cfg)

	// 2. Logger
	do.ProvideValue(injector, log)

	// 3. Database Connection
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

	// 4. Conditionally register domain-specific dependencies
	HealthRegistrationHook(injector)
	UserRegistrationHook(injector)
	PostRegistrationHook(injector)

	return injector, nil
}
