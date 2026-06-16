// gRPC server construction and interceptors.
package server

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

// NewGRPC builds and returns a *grpc.Server with OTel StatsHandler and a panic
// recovery/logging unary interceptor.
func NewGRPC(logger *zerolog.Logger) *grpc.Server {
	return grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(unaryInterceptor(logger)),
	)
}

// unaryInterceptor logs incoming gRPC requests and recovers from panics.
func unaryInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error().
					Str("method", info.FullMethod).
					Interface("panic", r).
					Msg("grpc panic recovered")

				err = sharederrors.GRPCErr(sharederrors.ErrInternal)
			}
		}()

		logger.Info().
			Str("method", info.FullMethod).
			Msg("grpc request")

		return handler(ctx, req)
	}
}
