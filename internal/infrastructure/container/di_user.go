//go:build user || all

package container

import (
	"database/sql"
	"github.com/samber/do/v2"
	"github.com/zercle/zercle-go-template/internal/core/port"
	userService "github.com/zercle/zercle-go-template/internal/features/user/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	userRepo "github.com/zercle/zercle-go-template/internal/features/user/repository"
	userHandler "github.com/zercle/zercle-go-template/internal/features/user/handler"
)

// RegisterUserHandler registers user-related dependencies
func RegisterUserHandler(i do.Injector) {
	// User Repository
	do.Provide(i, func(injector do.Injector) (port.UserRepository, error) {
		db := do.MustInvoke[*sql.DB](injector)
		return userRepo.NewUserRepository(db), nil
	})

	// User Service
	do.Provide(i, func(injector do.Injector) (port.UserService, error) {
		repo := do.MustInvoke[port.UserRepository](injector)
		cfg := do.MustInvoke[*config.Config](injector)
		return userService.NewUserService(repo, cfg.JWT.Secret, cfg.JWT.Expiration), nil
	})

	// User Handler
	do.Provide(i, func(injector do.Injector) (*userHandler.UserHandler, error) {
		svc := do.MustInvoke[port.UserService](injector)
		return userHandler.NewUserHandler(svc), nil
	})
}

// UserRegistrationHook is called from NewContainer
func UserRegistrationHook(i do.Injector) {
	RegisterUserHandler(i)
}
