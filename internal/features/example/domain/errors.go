// STUB FEATURE — delete internal/features/example to start your project.

package domain

import "errors"

// Domain sentinel errors for the example feature.
var (
	ErrItemNotFound = errors.New("item not found")
	ErrInvalidName  = errors.New("item name is invalid")
)
