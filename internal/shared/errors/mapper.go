// HTTP and gRPC mapping logic for the shared boundary errors.
package errors

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/status"
)

// HTTPError maps any error to an HTTP status code and a JSON-shaped response
// body. A nil error maps to 200 with a success body.
func HTTPError(err error) (int, map[string]any) {
	if err == nil {
		return http.StatusOK, map[string]any{"status": "ok"}
	}

	app := resolveAppError(err)

	body := map[string]any{
		"error":   app.Code,
		"message": app.Message,
	}
	if app.Cause != nil {
		body["cause"] = app.Cause.Error()
	}

	return app.HTTPStatus, body
}

// GRPCErr maps any error to a gRPC status error. A nil error maps to nil.
func GRPCErr(err error) error {
	if err == nil {
		return nil
	}

	app := resolveAppError(err)

	return fmt.Errorf("%w", status.Error(app.GRPCCode, app.Message))
}

// resolveAppError converts err into an AppError using, in order:
//  1. direct *AppError match via errors.As,
//  2. a registered domain sentinel via errors.Is,
//  3. the shared ErrInternal as a fallback.
func resolveAppError(err error) *AppError {
	var app *AppError
	if errors.As(err, &app) {
		return app
	}

	if app := sentinelFor(err); app != nil {
		return app
	}

	return ErrInternal
}
