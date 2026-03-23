package app

import (
	"github.com/samber/do/v2"

	task_usecase "github.com/zercle/zercle-go-template/internal/feature/task/usecase"
	"github.com/zercle/zercle-go-template/internal/feature/user"
	"github.com/zercle/zercle-go-template/internal/feature/user/usecase"
)

// RegisterServices registers all application service providers.
func RegisterServices(i do.Injector) {
	// User Service - depends on repository
	do.Provide(i, func(i do.Injector) (user.Usecase, error) {
		repo := do.MustInvoke[user.Repository](i)
		return usecase.NewService(repo), nil
	})

	// Task Service - depends on repository
	do.Provide(i, task_usecase.NewService)
}
