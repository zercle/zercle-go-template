package domain

import "errors"

// Domain sentinel errors for the machines feature. They are mapped to HTTP
// responses by the infrastructure error mapper, so their identity is part of
// the feature's public surface.
var (
	ErrMachineNotFound     = errors.New("machine not found")
	ErrInvalidID           = errors.New("machine id is invalid")
	ErrInvalidMachineLabel = errors.New("machine label is invalid")
	ErrUnsupportedCoin     = errors.New("unsupported coin")
)
