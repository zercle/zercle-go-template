package container

import (
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	httpAdapter "github.com/zercle/zercle-go-template/internal/adapter/handler/http"
	"github.com/zercle/zercle-go-template/internal/core/port"
	"github.com/zercle/zercle-go-template/internal/core/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// NewContainer initializes the DI container with common services.
// Repositories must be provided by the caller (Production or Test).
func NewContainer(cfg *config.Config, log zerolog.Logger) do.Injector {
	injector := do.New()

	// 1. Config
	do.ProvideValue(injector, cfg)

	// 2. Logger
	do.ProvideValue(injector, log)

	// 3. Services (Core Layer)
	do.Provide(injector, func(i do.Injector) (port.UserService, error) {
		repo := do.MustInvoke[port.UserRepository](i)
		cfg := do.MustInvoke[*config.Config](i)
		return service.NewUserService(repo, cfg.JWT.Secret, cfg.JWT.Expiration), nil
	})
	do.Provide(injector, func(i do.Injector) (port.PostService, error) {
		repo := do.MustInvoke[port.PostRepository](i)
		return service.NewPostService(repo), nil
	})

	// 4. Handlers (Adapter Layer)
	do.Provide(injector, func(i do.Injector) (*httpAdapter.UserHandler, error) {
		svc := do.MustInvoke[port.UserService](i)
		return httpAdapter.NewUserHandler(svc), nil
	})
	do.Provide(injector, func(i do.Injector) (*httpAdapter.PostHandler, error) {
		svc := do.MustInvoke[port.PostService](i)
		return httpAdapter.NewPostHandler(svc), nil
	})

	return injector
}
