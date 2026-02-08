// Package container provides dependency injection container management.
// It initializes and wires all application dependencies together.
package container

import (
	"fmt"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/feature/auth/usecase"
	userrepository "zercle-go-template/internal/feature/user/repository"
	userusecase "zercle-go-template/internal/feature/user/usecase"
	"zercle-go-template/internal/infrastructure/db"
	"zercle-go-template/internal/logger"
)

// Container holds all application dependencies.
// It provides a centralized location for accessing services and repositories.
type Container struct {
	Config      *config.Config
	Logger      logger.Logger
	UserRepo    userrepository.UserRepository
	UserUsecase userusecase.UserUsecase
	JWTUsecase  usecase.JWTUsecase
	db          *db.DB
}

// ContainerOption is a functional option for container configuration.
type ContainerOption func(*Container) error

// WithMemoryRepository configures the container to use the in-memory user repository.
// This is the default and is suitable for development and testing.
func WithMemoryRepository() ContainerOption {
	return func(c *Container) error {
		c.UserRepo = userrepository.NewMemoryUserRepository()
		c.Logger.Info("using in-memory user repository")
		return nil
	}
}

// WithSQLCRepository configures the container to use the sqlc-based PostgreSQL repository.
// This requires a valid database configuration and connection.
func WithSQLCRepository(database *db.DB) ContainerOption {
	return func(c *Container) error {
		if database == nil || database.Querier() == nil {
			return fmt.Errorf("database connection is required for SQLC repository")
		}
		c.UserRepo = userrepository.NewSqlcUserRepository(database.Querier())
		c.Logger.Info("using sqlc user repository")
		return nil
	}
}

// New creates a new dependency injection container.
// It initializes all dependencies based on the provided configuration.
// By default, it uses the in-memory repository. Use WithSQLCRepository for production.
func New(cfg *config.Config, opts ...ContainerOption) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	// Initialize logger
	log := logger.New(cfg.App.Name, cfg.App.Environment)

	// Initialize database connection (may be nil if not configured)
	database, err := db.New(cfg, log)
	if err != nil {
		log.Error("failed to initialize database, falling back to in-memory", logger.Error(err))
	}

	container := &Container{
		Config: cfg,
		Logger: log,
		db:     database,
	}

	// Apply default option if none provided
	if len(opts) == 0 {
		// Auto-select repository based on database availability
		if database != nil {
			opts = append(opts, WithSQLCRepository(database))
		} else {
			opts = append(opts, WithMemoryRepository())
		}
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(container); err != nil {
			return nil, fmt.Errorf("failed to apply container option: %w", err)
		}
	}

	// Initialize usecases with the selected repository
	container.UserUsecase = userusecase.NewUserUsecase(container.UserRepo, log)
	container.JWTUsecase = usecase.NewJWTUsecase(&cfg.JWT, log)

	log.Info("dependency container initialized successfully",
		logger.String("environment", cfg.App.Environment),
		logger.String("version", cfg.App.Version),
	)

	return container, nil
}

// Close cleans up resources held by the container.
func (c *Container) Close() error {
	if c.Logger != nil {
		c.Logger.Info("shutting down dependency container")
	}
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			c.Logger.Error("error closing database connection", logger.Error(err))
		}
	}
	return nil
}

// DB returns the database connection (may be nil if using in-memory).
func (c *Container) DB() *db.DB {
	return c.db
}

// UseMemoryRepository is a helper that returns true if the container is using the in-memory repository.
func (c *Container) UseMemoryRepository() bool {
	_, ok := c.UserRepo.(*userrepository.MemoryUserRepository)
	return ok
}

// UseSQLCRepository is a helper that returns true if the container is using the sqlc repository.
func (c *Container) UseSQLCRepository() bool {
	_, ok := c.UserRepo.(*userrepository.SqlcUserRepository)
	return ok
}
