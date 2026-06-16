//go:build unit

package errors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

var errDomainSentinel = errors.New("domain: widget not found")

func init() {
	sharederrors.RegisterSentinel(errDomainSentinel, sharederrors.ErrNotFound)
}

func TestHTTPError_Nil(t *testing.T) {
	status, body := sharederrors.HTTPError(nil)
	if status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected ok body, got %v", body)
	}
}

func TestHTTPError_AppError(t *testing.T) {
	app := &sharederrors.AppError{
		Code:       "BOOM",
		Message:    "boom message",
		HTTPStatus: http.StatusTeapot,
		GRPCCode:   codes.Unavailable,
		Cause:      errors.New("cause"),
	}
	status, body := sharederrors.HTTPError(app)
	if status != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, status)
	}
	if body["error"] != "BOOM" {
		t.Fatalf("expected code BOOM, got %v", body["error"])
	}
	if body["message"] != "boom message" {
		t.Fatalf("expected message, got %v", body["message"])
	}
	if _, ok := body["cause"]; ok {
		t.Fatal("cause must not leak")
	}
}

func TestHTTPError_RegisteredSentinel(t *testing.T) {
	status, body := sharederrors.HTTPError(errDomainSentinel)
	if status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, status)
	}
	if body["error"] != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %v", body["error"])
	}
}

func TestSentinelCausePreserved(t *testing.T) {
	wrapped := fmt.Errorf("wrap: %w", errDomainSentinel)
	status, body := sharederrors.HTTPError(wrapped)
	if status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, status)
	}
	if body["error"] != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %v", body["error"])
	}
}

func TestHTTPError_Unknown(t *testing.T) {
	status, body := sharederrors.HTTPError(errors.New("something went wrong"))
	if status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, status)
	}
	if body["error"] != "INTERNAL" {
		t.Fatalf("expected INTERNAL, got %v", body["error"])
	}
}

func TestHTTPError_UnknownDoesNotLeakCause(t *testing.T) {
	status, body := sharederrors.HTTPError(errors.New("secret internal: db query failed"))
	if status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, status)
	}
	if _, ok := body["cause"]; ok {
		t.Fatal("cause must not leak")
	}
	if body["message"] != "internal error" {
		t.Fatalf("expected sentinel message, got %v", body["message"])
	}
}

func TestGRPCErr_Nil(t *testing.T) {
	if sharederrors.GRPCErr(nil) != nil {
		t.Fatal("expected nil for nil error")
	}
}

func TestGRPCErr_AppError(t *testing.T) {
	app := &sharederrors.AppError{
		Code:       "BOOM",
		Message:    "boom message",
		HTTPStatus: http.StatusTeapot,
		GRPCCode:   codes.Unavailable,
	}
	err := sharederrors.GRPCErr(app)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected status error")
	}
	if st.Code() != codes.Unavailable {
		t.Fatalf("expected code %v, got %v", codes.Unavailable, st.Code())
	}
}

func TestGRPCErr_RegisteredSentinel(t *testing.T) {
	err := sharederrors.GRPCErr(errDomainSentinel)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected status error")
	}
	if st.Code() != codes.NotFound {
		t.Fatalf("expected code %v, got %v", codes.NotFound, st.Code())
	}
}

func TestGRPCErr_Unknown(t *testing.T) {
	err := sharederrors.GRPCErr(errors.New("random failure"))
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected status error")
	}
	if st.Code() != codes.Internal {
		t.Fatalf("expected code %v, got %v", codes.Internal, st.Code())
	}
}
