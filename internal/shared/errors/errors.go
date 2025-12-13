package sharederrors

import "errors"

// Domain errors
var (
	ErrNotFound       = errors.New("resource not found")
	ErrDuplicate      = errors.New("resource already exists")
	ErrInvalidCreds   = errors.New("invalid credentials")
	ErrInternalServer = errors.New("internal server error")
)
