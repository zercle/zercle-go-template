package domain

import "errors"

// Domain sentinel errors for the sales feature. They are mapped to HTTP
// responses by the infrastructure error mapper, so their identity is part of
// the feature's public surface.
var (
	ErrProductNotFound     = errors.New("product not found")
	ErrMachineNotFound     = errors.New("machine not found")
	ErrInvalidID           = errors.New("sale id is invalid")
	ErrUnsupportedCoin     = errors.New("unsupported coin")
	ErrInsufficientPayment = errors.New("insufficient payment")
	ErrOutOfStock          = errors.New("product out of stock")
	ErrExactChangeRequired = errors.New("exact change required")
)
