//go:build unit

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsClientSideCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code codes.Code
		want bool
	}{
		{codes.OK, false},
		{codes.Canceled, true},
		{codes.Unknown, false},
		{codes.InvalidArgument, true},
		{codes.DeadlineExceeded, true},
		{codes.NotFound, true},
		{codes.AlreadyExists, true},
		{codes.PermissionDenied, true},
		{codes.ResourceExhausted, true},
		{codes.FailedPrecondition, true},
		{codes.Aborted, true},
		{codes.OutOfRange, true},
		{codes.Unimplemented, false},
		{codes.Internal, false},
		{codes.Unavailable, false},
		{codes.DataLoss, false},
		{codes.Unauthenticated, true},
	}

	for _, tc := range tests {
		t.Run(tc.code.String(), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, isClientSideCode(tc.code))
		})
	}
}

func findCompletionLevel(t *testing.T, buf *bytes.Buffer) string {
	t.Helper()
	a := assert.New(t)
	for _, line := range bytes.Split(bytes.TrimRight(buf.Bytes(), "\n"), []byte("\n")) {
		if !bytes.Contains(line, []byte("grpc request completed")) {
			continue
		}
		var entry map[string]any
		a.NoError(json.Unmarshal(line, &entry))
		level, ok := entry["level"].(string)
		a.True(ok, "level field missing or not a string in line: %s", string(line))
		return level
	}
	t.Fatalf("no completion log line found in buffer: %s", buf.String())
	return ""
}

func TestUnaryInterceptorLogLevel(t *testing.T) {
	t.Parallel()

	newLogger := func() (*zerolog.Logger, *bytes.Buffer) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)
		return &logger, &buf
	}

	t.Run("success logs at info level", func(t *testing.T) {
		t.Parallel()
		logger, buf := newLogger()
		interceptor := unaryInterceptor(logger)
		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		handler := func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.NoError(t, err)
		assert.Equal(t, "ok", resp)
		assert.Equal(t, "info", findCompletionLevel(t, buf))
	})

	t.Run("client-side error logs at warn level", func(t *testing.T) {
		t.Parallel()
		logger, buf := newLogger()
		interceptor := unaryInterceptor(logger)
		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, status.Error(codes.NotFound, "missing")
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, "warn", findCompletionLevel(t, buf))
	})

	t.Run("server-side error logs at error level", func(t *testing.T) {
		t.Parallel()
		logger, buf := newLogger()
		interceptor := unaryInterceptor(logger)
		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, status.Error(codes.Internal, "boom")
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, "error", findCompletionLevel(t, buf))
	})

	t.Run("non-grpc error logs at error level", func(t *testing.T) {
		t.Parallel()
		logger, buf := newLogger()
		interceptor := unaryInterceptor(logger)
		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, errors.New("plain")
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, "error", findCompletionLevel(t, buf))
	})
}
