package di

import (
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
	"github.com/zercle/zercle-go-template/internal/features/auth/handler/http"
	"github.com/zercle/zercle-go-template/internal/features/auth/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
)

// ProvideUserRepository creates and provides a user repository.
func ProvideUserRepository(i do.Injector) (domain.UserRepository, error) {
	db := do.MustInvoke[*postgres.DB](i)
	return postgres.NewUserRepository(db), nil
}

// ProvideSessionRepository creates and provides a session repository.
func ProvideSessionRepository(i do.Injector) (domain.SessionRepository, error) {
	db := do.MustInvoke[*postgres.DB](i)
	return postgres.NewSessionRepository(db), nil
}

// ProvideAuthService creates and provides the auth service.
func ProvideAuthService(i do.Injector) (service.AuthServiceInterface, error) {
	cfg := do.MustInvoke[*config.Config](i)
	userRepo := do.MustInvoke[domain.UserRepository](i)
	sessionRepo := do.MustInvoke[domain.SessionRepository](i)

	return service.NewAuthService(
		userRepo,
		sessionRepo,
		cfg.AuthAccessTokenSecret,
		cfg.AuthAccessTokenTTL,
		cfg.AuthRefreshTokenTTL,
	), nil
}

// ProvideAuthHandler creates and provides the auth HTTP handler.
func ProvideAuthHandler(i do.Injector) (*http.AuthHandler, error) {
	authSvc := do.MustInvoke[service.AuthServiceInterface](i)
	return http.NewAuthHandler(authSvc), nil
}

// RegisterAuthProviders registers all auth-related DI providers.
func RegisterAuthProviders(i do.Injector) {
	do.Provide(i, ProvideUserRepository)
	do.Provide(i, ProvideSessionRepository)
	do.Provide(i, ProvideAuthService)
	do.Provide(i, ProvideAuthHandler)
}
