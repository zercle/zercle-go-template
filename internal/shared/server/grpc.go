// gRPC server construction and interceptors.
package server

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

// NewGRPC builds and returns a *grpc.Server with OTel StatsHandler, panic
// recovery + logging interceptors (unary and stream), and conservative
// message-size limits.
func NewGRPC(logger *zerolog.Logger) *grpc.Server {
	return grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(unaryInterceptor(logger)),
		grpc.StreamInterceptor(streamInterceptor(logger)),
		grpc.MaxRecvMsgSize(4*1024*1024),
		grpc.MaxSendMsgSize(4*1024*1024),
	)
}

// unaryInterceptor logs incoming gRPC requests, recovers from panics, and logs
// request completion with latency and resulting gRPC status.
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

		start := time.Now()
		resp, err = handler(ctx, req)
		latency := time.Since(start)
		if err != nil {
			logger.Error().
				Str("method", info.FullMethod).
				Dur("latency", latency).
				Err(err).
				Msg("grpc request completed")
		} else {
			logger.Info().
				Str("method", info.FullMethod).
				Dur("latency", latency).
				Msg("grpc request completed")
		}
		return resp, err
	}
}

// streamInterceptor mirrors unaryInterceptor's recovery logic for streaming
// RPCs so a panic in a stream handler is logged and converted to a gRPC
// internal error rather than crashing the server.
func streamInterceptor(logger *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				if recErr, ok := r.(error); ok {
					logger.Error().Err(recErr).Str("method", info.FullMethod).Msg("grpc stream panic recovered")
				} else {
					logger.Error().Interface("panic", r).Str("method", info.FullMethod).Msg("grpc stream panic recovered")
				}
				err = sharederrors.GRPCErr(sharederrors.ErrInternal)
			}
		}()
		return handler(srv, ss)
	}
}
