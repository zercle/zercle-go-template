package di

import (
	"github.com/samber/do"
	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
	"github.com/zercle/zercle-go-template/internal/features/auth/handler/http"
	"github.com/zercle/zercle-go-template/internal/features/auth/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
)

func ProvideUserRepository(i *do.Injector) (domain.UserRepository, error) {
	db := do.MustInvoke[*postgres.DB](i)
	return postgres.NewUserRepository(db), nil
}

func ProvideSessionRepository(i *do.Injector) (domain.SessionRepository, error) {
	db := do.MustInvoke[*postgres.DB](i)
	return postgres.NewSessionRepository(db), nil
}

func ProvideAuthService(i *do.Injector) (service.AuthServiceInterface, error) {
	cfg := do.MustInvoke[*config.Config](i)
	userRepo := do.MustInvoke[domain.UserRepository](i)
	sessionRepo := do.MustInvoke[domain.SessionRepository](i)

	return service.NewAuthService(
		userRepo,
		sessionRepo,
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTExpiry,
		cfg.Auth.RefreshExpiry,
	), nil
}

func ProvideAuthHandler(i *do.Injector) (*http.AuthHandler, error) {
	authSvc := do.MustInvoke[service.AuthServiceInterface](i)
	return http.NewAuthHandler(authSvc), nil
}

func RegisterAuthProviders(i *do.Injector) {
	do.Provide(i, ProvideUserRepository)
	do.Provide(i, ProvideSessionRepository)
	do.Provide(i, ProvideAuthService)
	do.Provide(i, ProvideAuthHandler)
}
