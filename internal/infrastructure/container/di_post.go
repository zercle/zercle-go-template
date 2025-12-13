//go:build post || all

package container

import (
	"database/sql"
	"github.com/samber/do/v2"
	"github.com/zercle/zercle-go-template/internal/core/port"
	postService "github.com/zercle/zercle-go-template/internal/features/post/service"
	postRepo "github.com/zercle/zercle-go-template/internal/features/post/repository"
	postHandler "github.com/zercle/zercle-go-template/internal/features/post/handler"
)

// RegisterPostHandler registers post-related dependencies
func RegisterPostHandler(i do.Injector) {
	// Post Repository
	do.Provide(i, func(injector do.Injector) (port.PostRepository, error) {
		db := do.MustInvoke[*sql.DB](injector)
		return postRepo.NewPostRepository(db), nil
	})

	// Post Service
	do.Provide(i, func(injector do.Injector) (port.PostService, error) {
		repo := do.MustInvoke[port.PostRepository](injector)
		return postService.NewPostService(repo), nil
	})

	// Post Handler
	do.Provide(i, func(injector do.Injector) (*postHandler.PostHandler, error) {
		svc := do.MustInvoke[port.PostService](injector)
		return postHandler.NewPostHandler(svc), nil
	})
}

// PostRegistrationHook is called from NewContainer
func PostRegistrationHook(i do.Injector) {
	RegisterPostHandler(i)
}
