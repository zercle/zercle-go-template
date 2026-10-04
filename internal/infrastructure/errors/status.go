// HTTP status to boundary-error mapping, so framework-generated errors (router
// 404/405, body-limit 413) are served in the shared envelope with the correct
// code instead of being collapsed to INTERNAL.
package errors

import (
	"net/http"

	"github.com/zercle/zercle-go-template/pkg/api/errcodes"
)

// Sentinel mappings for framework statuses that have no domain sentinel of
// their own. They are values, not exported vars, because callers receive clones.
var (
	errMethodNotAllowed = &AppError{Code: errcodes.MethodNotAllowed, Message: "method not allowed", HTTPStatus: http.StatusMethodNotAllowed}
	errPayloadTooLarge  = &AppError{Code: errcodes.PayloadTooLarge, Message: "payload too large", HTTPStatus: http.StatusRequestEntityTooLarge}
)

// ForHTTPStatus maps an HTTP status code to the shared AppError that best
// represents it. It exists for errors produced by the HTTP framework itself —
// echo's router emits *echo.HTTPError values for 404/405 and its BodyLimit
// middleware for 413 — which carry a status but no domain meaning. The returned
// AppError is always a fresh clone, so callers may not mutate shared sentinels.
func ForHTTPStatus(status int) *AppError {
	switch status {
	case http.StatusBadRequest:
		return clone(ErrInvalidInput)
	case http.StatusUnauthorized:
		return clone(ErrUnauthorized)
	case http.StatusForbidden:
		return clone(ErrForbidden)
	case http.StatusNotFound:
		return clone(ErrNotFound)
	case http.StatusMethodNotAllowed:
		return clone(errMethodNotAllowed)
	case http.StatusConflict:
		return clone(ErrConflict)
	case http.StatusRequestEntityTooLarge:
		return clone(errPayloadTooLarge)
	case 499: // client closed request (nginx convention, matches ErrCanceled)
		return clone(ErrCanceled)
	case http.StatusGatewayTimeout:
		return clone(ErrDeadlineExceeded)
	}

	// Unknown statuses keep their status but borrow the class's code, so a
	// client buffering on our generic codes still behaves sensibly.
	if status >= http.StatusInternalServerError {
		app := clone(ErrInternal)
		app.HTTPStatus = status
		return app
	}

	app := clone(ErrInvalidInput)
	app.HTTPStatus = status
	return app
}

// clone returns a copy of app so callers can never mutate a shared sentinel.
func clone(app *AppError) *AppError {
	c := *app
	return &c
}

// withCause returns a clone of app carrying err as its cause, keeping the
// wrapped error available for observability without mutating the sentinel.
func withCause(app *AppError, err error) *AppError {
	c := clone(app)
	c.Cause = err
	return c
}
