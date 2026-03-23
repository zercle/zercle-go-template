package app

import (
	"github.com/samber/do/v2"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	task_handler "github.com/zercle/zercle-go-template/internal/feature/task/handler"
	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
	"github.com/zercle/zercle-go-template/internal/feature/user/handler"
)

// RegisterHandlers registers all HTTP handler providers.
func RegisterHandlers(i do.Injector) {
	// User Handler - depends on user service
	do.Provide(i, func(i do.Injector) (user_entity.Handler, error) {
		svc := do.MustInvoke[user_entity.Usecase](i)
		return handler.NewHandler(svc), nil
	})

	// Task Handler - depends on task service
	do.Provide(i, func(i do.Injector) (task_entity.Handler, error) {
		return task_handler.NewHandler(i)
	})
}
