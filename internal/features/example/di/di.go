// STUB FEATURE — delete internal/features/example to start your project.

package di

import (
	"fmt"

	"github.com/samber/do/v2"

	pb "github.com/zercle/zercle-go-template/api/pb/example/v1"
	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	grpchandler "github.com/zercle/zercle-go-template/internal/features/example/handler/grpc"
	httphandler "github.com/zercle/zercle-go-template/internal/features/example/handler/http"
	"github.com/zercle/zercle-go-template/internal/features/example/repository"
	"github.com/zercle/zercle-go-template/internal/features/example/service"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Register wires the example feature into the composition root.
func Register(c do.Injector) error {
	sharederrors.RegisterSentinel(domain.ErrItemNotFound, sharederrors.ErrNotFound)
	sharederrors.RegisterSentinel(domain.ErrInvalidName, sharederrors.ErrInvalidInput)
	sharederrors.RegisterSentinel(domain.ErrInvalidID, sharederrors.ErrInvalidInput)

	do.Provide(c, func(i do.Injector) (domain.Repository, error) {
		gormDB, err := do.Invoke[*gorm.DB](i)
		if err != nil {
			return nil, fmt.Errorf("resolve gorm db: %w", err)
		}
		return repository.NewRepository(gormDB), nil
	})

	do.Provide(c, func(i do.Injector) (domain.Service, error) {
		repo, err := do.Invoke[domain.Repository](i)
		if err != nil {
			return nil, fmt.Errorf("resolve example repository: %w", err)
		}
		cfg, err := do.Invoke[*config.Config](i)
		if err != nil {
			return nil, fmt.Errorf("resolve config: %w", err)
		}
		return service.NewService(repo, cfg.Example.DefaultPageSize, cfg.Example.MaxPageSize, cfg.Example.MaxNameLength), nil
	})

	do.Provide(c, func(i do.Injector) (*httphandler.Handler, error) {
		svc, err := do.Invoke[domain.Service](i)
		if err != nil {
			return nil, fmt.Errorf("resolve example service: %w", err)
		}
		return httphandler.New(svc), nil
	})

	do.Provide(c, func(i do.Injector) (*grpchandler.Server, error) {
		svc, err := do.Invoke[domain.Service](i)
		if err != nil {
			return nil, fmt.Errorf("resolve example service: %w", err)
		}
		return grpchandler.NewServer(svc), nil
	})

	h, err := do.Invoke[*httphandler.Handler](c)
	if err != nil {
		return fmt.Errorf("resolve example http handler: %w", err)
	}
	e, err := do.Invoke[*echo.Echo](c)
	if err != nil {
		return fmt.Errorf("resolve example echo: %w", err)
	}
	g := e.Group("/api/v1")
	h.Register(g)

	gs, err := do.Invoke[*grpc.Server](c)
	if err != nil {
		return fmt.Errorf("resolve example grpc server: %w", err)
	}
	grpcHandler, err := do.Invoke[*grpchandler.Server](c)
	if err != nil {
		return fmt.Errorf("resolve example grpc handler: %w", err)
	}
	pb.RegisterExampleServiceServer(gs, grpcHandler)

	return nil
}
