package apiv1

import "github.com/zercle/zercle-go-template/pkg/api/errcodes"

// Error codes carried in the {"error": code, "message": msg} response
// envelope, re-exported from the version-independent errcodes package.
const (
	ErrCodeNotFound         = errcodes.NotFound
	ErrCodeInvalidInput     = errcodes.InvalidInput
	ErrCodeUnauthorized     = errcodes.Unauthorized
	ErrCodeForbidden        = errcodes.Forbidden
	ErrCodeConflict         = errcodes.Conflict
	ErrCodeCanceled         = errcodes.Canceled
	ErrCodeDeadlineExceeded = errcodes.DeadlineExceeded
	ErrCodeInternal         = errcodes.Internal
	ErrCodeMethodNotAllowed = errcodes.MethodNotAllowed
	ErrCodePayloadTooLarge  = errcodes.PayloadTooLarge
)
