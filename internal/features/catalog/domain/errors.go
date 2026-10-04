package domain

import "errors"

// Domain sentinel errors for the catalog feature. They are mapped to HTTP
// responses by the infrastructure error mapper, so their identity is part of
// the feature's public surface.
var (
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidID          = errors.New("product id is invalid")
	ErrInvalidProductName = errors.New("product name is invalid")
	ErrInvalidPrice       = errors.New("product price is invalid")
)
