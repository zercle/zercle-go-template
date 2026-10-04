// Package di wires the sales feature into the composition root.
package di

import (
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/handler"
	"github.com/zercle/zercle-go-template/internal/features/sales/repository"
	"github.com/zercle/zercle-go-template/internal/features/sales/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/features/sales/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// Register wires the sales feature into the composition root. When
// cfg.Sales.Enabled is false the feature is not registered at all: no
// providers, no HTTP routes, no sentinel mappings.
func Register(c do.Injector) error {
	cfg, err := do.Invoke[*config.Config](c)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}
	if !cfg.Sales.Enabled {
		return nil
	}

	apperrors.RegisterSentinel(domain.ErrProductNotFound, apperrors.ErrNotFound)
	apperrors.RegisterSentinel(domain.ErrMachineNotFound, apperrors.ErrNotFound)
	apperrors.RegisterSentinel(domain.ErrOutOfStock, apperrors.ErrConflict)
	apperrors.RegisterSentinel(domain.ErrInvalidID, apperrors.ErrInvalidInput)
	apperrors.RegisterSentinel(domain.ErrUnsupportedCoin, apperrors.ErrInvalidInput)
	apperrors.RegisterSentinel(domain.ErrInsufficientPayment, apperrors.ErrInvalidInput)
	apperrors.RegisterSentinel(domain.ErrExactChangeRequired, apperrors.ErrInvalidInput)

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
			return nil, fmt.Errorf("resolve sales repository: %w", err)
		}
		return usecase.NewUsecase(repo), nil
	})

	do.Provide(c, func(i do.Injector) (*handler.Handler, error) {
		svc, err := do.Invoke[usecase.Service](i)
		if err != nil {
			return nil, fmt.Errorf("resolve sales service: %w", err)
		}
		return handler.New(svc), nil
	})

	h, err := do.Invoke[*handler.Handler](c)
	if err != nil {
		return fmt.Errorf("resolve sales http handler: %w", err)
	}
	e, err := do.Invoke[*echo.Echo](c)
	if err != nil {
		return fmt.Errorf("resolve sales echo: %w", err)
	}
	g := e.Group("/api/v1")
	h.Register(g)

	return nil
}
