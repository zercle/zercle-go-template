// HTTP mapping logic for the shared boundary errors.
package errors

import (
	"context"
	"errors"
)

// HTTPError maps any error to an HTTP status code and a JSON-shaped error body
// ({"error": code, "message": msg}). A nil error is programmer error: there is
// no failure to describe, so it returns (0, nil) rather than fabricating a 200
// "success" envelope that no caller can distinguish from a real response.
func HTTPError(err error) (int, map[string]any) {
	if err == nil {
		return 0, nil
	}

	app := resolveAppError(err)

	body := map[string]any{
		"error":   app.Code,
		"message": app.Message,
	}

	return app.HTTPStatus, body
}

// resolveAppError converts err into an AppError using, in order:
//  1. direct *AppError match via errors.As,
//  2. a registered domain sentinel via errors.Is,
//  3. the standard context errors (Canceled, DeadlineExceeded),
//  4. the shared ErrInternal as a fallback.
//
// Every successful path returns a clone of the matched AppError so callers can
// never mutate shared sentinels or the AppError they passed in.
func resolveAppError(err error) *AppError {
	if app, ok := errors.AsType[*AppError](err); ok {
		return withCause(app, err)
	}

	if app := sentinelFor(err); app != nil {
		return withCause(app, err)
	}

	if errors.Is(err, context.Canceled) {
		return withCause(ErrCanceled, err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return withCause(ErrDeadlineExceeded, err)
	}

	return withCause(ErrInternal, err)
}
