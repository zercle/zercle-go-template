package domain

import "errors"

// Domain sentinel errors for the reporting feature. They are mapped to HTTP
// responses by the infrastructure error mapper, so their identity is part of
// the feature's public surface.
var (
	// ErrInvalidTopMachines reports a requested top-machines limit above the
	// configured maximum.
	ErrInvalidTopMachines = errors.New("top machines limit is invalid")
)
