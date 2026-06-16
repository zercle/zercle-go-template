// STUB FEATURE — delete internal/features/example to start your project.

package di

import (
	"github.com/samber/do/v2"

	pb "github.com/zercle/zercle-go-template/api/pb/example/v1"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	grpchandler "github.com/zercle/zercle-go-template/internal/features/example/handler/grpc"
	httphandler "github.com/zercle/zercle-go-template/internal/features/example/handler/http"
	"github.com/zercle/zercle-go-template/internal/features/example/repository"
	"github.com/zercle/zercle-go-template/internal/features/example/service"
	sqlcdb "github.com/zercle/zercle-go-template/internal/infrastructure/db/sqlc"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"google.golang.org/grpc"
)

// Register wires the example feature into the composition root.
func Register(c do.Injector) error {
	sharederrors.RegisterSentinel(domain.ErrItemNotFound, sharederrors.ErrNotFound)
	sharederrors.RegisterSentinel(domain.ErrInvalidName, sharederrors.ErrInvalidInput)

	do.Provide(c, func(i do.Injector) (domain.Repository, error) {
		pool := do.MustInvoke[*pgxpool.Pool](i)
		queries := do.MustInvoke[*sqlcdb.Queries](i)
		return repository.NewRepository(pool, queries), nil
	})

	do.Provide(c, func(i do.Injector) (domain.Service, error) {
		repo := do.MustInvoke[domain.Repository](i)
		return service.NewService(repo), nil
	})

	do.Provide(c, func(i do.Injector) (*httphandler.Handler, error) {
		svc := do.MustInvoke[domain.Service](i)
		return httphandler.New(svc), nil
	})

	do.Provide(c, func(i do.Injector) (*grpchandler.Server, error) {
		svc := do.MustInvoke[domain.Service](i)
		return grpchandler.NewServer(svc), nil
	})

	h, err := do.Invoke[*httphandler.Handler](c)
	if err != nil {
		return err
	}
	e, err := do.Invoke[*echo.Echo](c)
	if err != nil {
		return err
	}
	g := e.Group("/api/v1")
	h.Register(g)

	gs, err := do.Invoke[*grpc.Server](c)
	if err != nil {
		return err
	}
	grpcHandler, err := do.Invoke[*grpchandler.Server](c)
	if err != nil {
		return err
	}
	pb.RegisterExampleServiceServer(gs, grpcHandler)

	return nil
}
