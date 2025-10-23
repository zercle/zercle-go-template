package container

import (
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/zercle/zercle-go-template/internal/core/port"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"go.uber.org/mock/gomock"
)

// NewTestContainer initializes the DI container for Integration Tests.
// It bundles the common container with Mock Repositories.
// Callers must invoke do.Override to replace specific mocks if needed,
// but here we can provide a base set if we want (or leave empty to force override).
func NewTestContainer(cfg *config.Config, log zerolog.Logger, userRepo port.UserRepository, postRepo port.PostRepository) do.Injector {
	injector := NewContainer(cfg, log)

	// Register Mock Repositories
	do.Provide(injector, func(i do.Injector) (port.UserRepository, error) {
		return userRepo, nil
	})
	do.Provide(injector, func(i do.Injector) (port.PostRepository, error) {
		return postRepo, nil
	})

	return injector
}

// MockProvider defines a function that returns mock repositories.
type MockProvider func(ctrl *gomock.Controller) (port.UserRepository, port.PostRepository)

// NewMockedContainer creates a container with Mocks from a controller.
func NewMockedContainer(cfg *config.Config, log zerolog.Logger, mocks MockProvider, ctrl *gomock.Controller) do.Injector {
	userRepo, postRepo := mocks(ctrl)
	return NewTestContainer(cfg, log, userRepo, postRepo)
}
