//go:build health || all

package container

import (
	"database/sql"

	"github.com/samber/do/v2"
	"github.com/zercle/zercle-go-template/internal/core/port"
	healthRepo "github.com/zercle/zercle-go-template/internal/features/health/repository"
	healthService "github.com/zercle/zercle-go-template/internal/features/health/service"
)

// RegisterHealthHandler registers health-specific dependencies
func RegisterHealthHandler(i do.Injector) {
	// Register Health Repository
	do.Provide(i, func(i *do.Injector) port.HealthRepository {
		db := do.MustInvoke[*sql.DB](i)
		return healthRepo.NewHealthRepository(db)
	})

	// Register Health Service
	do.Provide(i, func(i *do.Injector) port.HealthService {
		repo := do.MustInvoke[port.HealthRepository](i)
		return healthService.NewHealthService(repo)
	})
}

// HealthRegistrationHook is called from NewContainer
func HealthRegistrationHook(i do.Injector) {
	RegisterHealthHandler(i)
}
