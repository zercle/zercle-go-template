// Package di wires the machines feature into the composition root.
package di

import (
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/handler"
	"github.com/zercle/zercle-go-template/internal/features/machines/repository"
	"github.com/zercle/zercle-go-template/internal/features/machines/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/features/machines/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// Register wires the machines feature into the composition root. When
// cfg.Machines.Enabled is false the feature is not registered at all: no
// providers, no HTTP routes, no sentinel mappings.
func Register(c do.Injector) error {
	cfg, err := do.Invoke[*config.Config](c)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}
	if !cfg.Machines.Enabled {
		return nil
	}

	apperrors.RegisterSentinel(domain.ErrMachineNotFound, apperrors.ErrNotFound)
	apperrors.RegisterSentinel(domain.ErrInvalidID, apperrors.ErrInvalidInput)
	apperrors.RegisterSentinel(domain.ErrInvalidMachineLabel, apperrors.ErrInvalidInput)
	apperrors.RegisterSentinel(domain.ErrUnsupportedCoin, apperrors.ErrInvalidInput)

	do.Provide(c, func(i do.Injector) (repository.Repository, error) {
		gormDB, err := do.Invoke[*gorm.DB](i)
		if err != nil {
			return nil, fmt.Errorf("resolve gorm db: %w", err)
		}
		return postgres.NewRepository(gormDB), nil
	})

	do.Provide(c, func(i do.Injector) (usecase.Service, error) {
		repo, err := do.Invoke[repository.Repository](i)
		if err != nil {
			return nil, fmt.Errorf("resolve machines repository: %w", err)
		}
		cfg, err := do.Invoke[*config.Config](i)
		if err != nil {
			return nil, fmt.Errorf("resolve config: %w", err)
		}
		return usecase.NewUsecase(repo, cfg.Machines.DefaultPageSize, cfg.Machines.MaxPageSize, cfg.Machines.MaxLabelLength), nil
	})

	do.Provide(c, func(i do.Injector) (*handler.Handler, error) {
		svc, err := do.Invoke[usecase.Service](i)
		if err != nil {
			return nil, fmt.Errorf("resolve machines service: %w", err)
		}
		return handler.New(svc), nil
	})

	h, err := do.Invoke[*handler.Handler](c)
	if err != nil {
		return fmt.Errorf("resolve machines http handler: %w", err)
	}
	e, err := do.Invoke[*echo.Echo](c)
	if err != nil {
		return fmt.Errorf("resolve machines echo: %w", err)
	}
	g := e.Group("/api/v1")
	h.Register(g)

	return nil
}
