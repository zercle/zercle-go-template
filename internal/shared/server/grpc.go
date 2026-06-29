// gRPC server construction and interceptors.
package server

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
				err = recoverGRPCPanic(logger, r, info.FullMethod, "unary")
			}
		}()

		logger.Info().
			Str("method", info.FullMethod).
			Msg("grpc request")

		start := time.Now()
		resp, err = handler(ctx, req)
		latency := time.Since(start)
		if err != nil {
			ev := logger.Error()
			if st, ok := status.FromError(err); ok {
				if isClientSideCode(st.Code()) {
					ev = logger.Warn()
				}
			}
			ev.
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

// streamInterceptor recovers from panics in streaming RPC handlers via the
// shared recoverGRPCPanic helper so a panic is logged and converted to a gRPC
// internal error rather than crashing the server.
func streamInterceptor(logger *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverGRPCPanic(logger, r, info.FullMethod, "stream")
			}
		}()
		return handler(srv, ss)
	}
}

// recoverGRPCPanic logs a recovered panic at Error level and returns a gRPC
// internal error. When the recovered value satisfies the error interface the
// underlying error is attached via zerolog's .Err() so the error chain is
// preserved; otherwise the raw value is logged under the "panic" field.
// `kind` labels the interceptor source ("unary" or "stream") in the log
// message. Returns nil when r is nil so callers can guard with a single
// `if r := recover(); r != nil { ... }`.
func recoverGRPCPanic(logger *zerolog.Logger, r any, method, kind string) error {
	if r == nil {
		return nil
	}
	ev := logger.Error().Str("method", method)
	if recErr, ok := r.(error); ok {
		ev = ev.Err(recErr)
	} else {
		ev = ev.Interface("panic", r)
	}
	ev.Msgf("grpc %s panic recovered", kind)
	return sharederrors.GRPCErr(sharederrors.ErrInternal)
}

// isClientSideCode reports whether a gRPC status code represents an
// expected client-side error that should not trigger server-side alerting.
// Server-side failures (Internal, Unknown, Unimplemented, DataLoss,
// Unavailable) return false so they remain at Error level.
func isClientSideCode(c codes.Code) bool {
	//nolint:exhaustive // Server-side codes (Internal, Unknown, Unimplemented, DataLoss, Unavailable) intentionally fall through to false.
	switch c {
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.PermissionDenied, codes.Unauthenticated, codes.Canceled,
		codes.DeadlineExceeded, codes.ResourceExhausted,
		codes.FailedPrecondition, codes.Aborted, codes.OutOfRange:
		return true
	}
	return false
}
