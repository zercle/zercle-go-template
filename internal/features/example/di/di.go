// STUB FEATURE — delete internal/features/example to start your project.

package di

import (
	"errors"
	"fmt"

	"github.com/samber/do/v2"
	"github.com/valkey-io/valkey-go/valkeyaside"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/handler"
	"github.com/zercle/zercle-go-template/internal/features/example/repository"
	"github.com/zercle/zercle-go-template/internal/features/example/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/features/example/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Register wires the example feature into the composition root. When
// cfg.Example.Enabled is false the feature is not registered at all: no
// providers, no HTTP routes, no sentinel mappings.
func Register(c do.Injector) error {
	cfg, err := do.Invoke[*config.Config](c)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}
	if !cfg.Example.Enabled {
		return nil
	}

	apperrors.RegisterSentinel(domain.ErrItemNotFound, apperrors.ErrNotFound)
	apperrors.RegisterSentinel(domain.ErrInvalidName, apperrors.ErrInvalidInput)
	apperrors.RegisterSentinel(domain.ErrInvalidID, apperrors.ErrInvalidInput)

	do.Provide(c, func(i do.Injector) (repository.Repository, error) {
		gormDB, err := do.Invoke[*gorm.DB](i)
		if err != nil {
			return nil, fmt.Errorf("resolve gorm db: %w", err)
		}
		repo := postgres.NewRepository(gormDB)

		// Decorate with cache-aside reads when a Valkey cache-aside client is
		// registered; otherwise the feature works directly against the database.
		// Only a missing registration falls back — a construction failure is
		// a real error and must not be swallowed.
		aside, err := do.Invoke[valkeyaside.CacheAsideClient](i)
		if errors.Is(err, do.ErrServiceNotFound) {
			return repo, nil
		}
		if err != nil {
			return nil, fmt.Errorf("resolve valkey cache-aside client: %w", err)
		}
		return postgres.NewCachedRepository(repo, aside, cfg.Valkey.TTL), nil
	})

	do.Provide(c, func(i do.Injector) (usecase.Service, error) {
		repo, err := do.Invoke[repository.Repository](i)
		if err != nil {
			return nil, fmt.Errorf("resolve example repository: %w", err)
		}
		cfg, err := do.Invoke[*config.Config](i)
		if err != nil {
			return nil, fmt.Errorf("resolve config: %w", err)
		}
		return usecase.NewUsecase(repo, cfg.Example.DefaultPageSize, cfg.Example.MaxPageSize, cfg.Example.MaxNameLength), nil
	})

	do.Provide(c, func(i do.Injector) (*handler.Handler, error) {
		svc, err := do.Invoke[usecase.Service](i)
		if err != nil {
			return nil, fmt.Errorf("resolve example service: %w", err)
		}
		return handler.New(svc), nil
	})

	h, err := do.Invoke[*handler.Handler](c)
	if err != nil {
		return fmt.Errorf("resolve example http handler: %w", err)
	}
	e, err := do.Invoke[*echo.Echo](c)
	if err != nil {
		return fmt.Errorf("resolve example echo: %w", err)
	}
	g := e.Group("/api/v1")
	h.Register(g)

	return nil
}
