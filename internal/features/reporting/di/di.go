// Package di wires the reporting feature into the composition root.
package di

import (
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	"github.com/zercle/zercle-go-template/internal/features/reporting/handler"
	"github.com/zercle/zercle-go-template/internal/features/reporting/repository"
	"github.com/zercle/zercle-go-template/internal/features/reporting/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/features/reporting/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// Register wires the reporting feature into the composition root. When
// cfg.Reporting.Enabled is false the feature is not registered at all: no
// providers, no HTTP routes, no sentinel mappings. The feature owns no schema,
// so it contributes no migrations.
func Register(c do.Injector) error {
	cfg, err := do.Invoke[*config.Config](c)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}
	if !cfg.Reporting.Enabled {
		return nil
	}

	apperrors.RegisterSentinel(domain.ErrInvalidTopMachines, apperrors.ErrInvalidInput)

	do.Provide(c, func(i do.Injector) (repository.Repository, error) {
		gormDB, err := do.Invoke[*gorm.DB](i)
		if err != nil {
			return nil, fmt.Errorf("resolve gorm db: %w", err)
		}
		return postgres.NewRepository(gormDB), nil
	})

	do.Provide(c, func(i do.Injector) (usecase.Service, error) {
		reportingRepo, err := do.Invoke[repository.Repository](i)
		if err != nil {
			return nil, fmt.Errorf("resolve reporting repository: %w", err)
		}
		cfg, err := do.Invoke[*config.Config](i)
		if err != nil {
			return nil, fmt.Errorf("resolve config: %w", err)
		}
		return usecase.NewUsecase(reportingRepo, cfg.Reporting.DefaultTopMachines, cfg.Reporting.MaxTopMachines), nil
	})

	do.Provide(c, func(i do.Injector) (*handler.Handler, error) {
		svc, err := do.Invoke[usecase.Service](i)
		if err != nil {
			return nil, fmt.Errorf("resolve reporting service: %w", err)
		}
		return handler.New(svc), nil
	})

	h, err := do.Invoke[*handler.Handler](c)
	if err != nil {
		return fmt.Errorf("resolve reporting http handler: %w", err)
	}
	e, err := do.Invoke[*echo.Echo](c)
	if err != nil {
		return fmt.Errorf("resolve reporting echo: %w", err)
	}
	g := e.Group("/api/v1")
	h.Register(g)

	return nil
}
