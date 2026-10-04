//go:build unit

package errors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	apperrors "github.com/zercle/zercle-go-template/internal/platform/errors"
)

var errDomainSentinel = errors.New("domain: widget not found")

func init() {
	apperrors.RegisterSentinel(errDomainSentinel, apperrors.ErrNotFound)
}

// TestHTTPError_Nil pins that a nil error is not mapped to a fabricated 200
// "success" envelope: there is no failure to describe, so the mapper returns a
// zero status and no body, forcing the caller to handle it explicitly.
func TestHTTPError_Nil(t *testing.T) {
	status, body := apperrors.HTTPError(nil)
	if status != 0 {
		t.Fatalf("expected zero status for nil error, got %d", status)
	}
	if body != nil {
		t.Fatalf("expected nil body for nil error, got %v", body)
	}
}

func TestHTTPError_AppError(t *testing.T) {
	app := &apperrors.AppError{
		Code:       "BOOM",
		Message:    "boom message",
		HTTPStatus: http.StatusTeapot,
		Cause:      errors.New("cause"),
	}
	status, body := apperrors.HTTPError(app)
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
	status, body := apperrors.HTTPError(errDomainSentinel)
	if status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, status)
	}
	if body["error"] != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %v", body["error"])
	}
}

func TestSentinelCausePreserved(t *testing.T) {
	wrapped := fmt.Errorf("wrap: %w", errDomainSentinel)
	status, body := apperrors.HTTPError(wrapped)
	if status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, status)
	}
	if body["error"] != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %v", body["error"])
	}
}

func TestHTTPError_Unknown(t *testing.T) {
	status, body := apperrors.HTTPError(errors.New("something went wrong"))
	if status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, status)
	}
	if body["error"] != "INTERNAL" {
		t.Fatalf("expected INTERNAL, got %v", body["error"])
	}
}

func TestHTTPError_UnknownDoesNotLeakCause(t *testing.T) {
	status, body := apperrors.HTTPError(errors.New("secret internal: db query failed"))
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
