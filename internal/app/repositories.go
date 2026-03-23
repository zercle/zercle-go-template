package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	task_repository "github.com/zercle/zercle-go-template/internal/feature/task/repository"
	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
	user_repository "github.com/zercle/zercle-go-template/internal/feature/user/repository"
)

// RegisterRepositories registers all repository providers.
func RegisterRepositories(i do.Injector) {
	// User Repository - depends on database pool
	do.Provide(i, func(i do.Injector) (user_entity.Repository, error) {
		pool := do.MustInvoke[*pgxpool.Pool](i)
		return user_repository.NewPostgresRepository(pool), nil
	})

	// Task Repository - depends on database pool
	do.Provide(i, func(i do.Injector) (task_entity.Repository, error) {
		pool := do.MustInvoke[*pgxpool.Pool](i)
		return task_repository.NewPostgresRepository(pool), nil
	})
}
