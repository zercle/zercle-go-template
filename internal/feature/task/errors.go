package task

import "errors"

// Domain errors for the task feature.
var (
	// ErrTaskNotFound indicates the requested task does not exist.
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskAlreadyExists indicates a task with the same identifier already exists.
	ErrTaskAlreadyExists = errors.New("task already exists")

	// ErrInvalidTaskStatus indicates the provided task status is invalid.
	ErrInvalidTaskStatus = errors.New("invalid task status")

	// ErrInvalidTaskPriority indicates the provided task priority is invalid.
	ErrInvalidTaskPriority = errors.New("invalid task priority")
)

// DomainError is a sentinel error for task domain errors that supports errors.Is comparison.
type DomainError struct {
	Sentinel error
}

func (e *DomainError) Error() string {
	return e.Sentinel.Error()
}

func (e *DomainError) Is(target error) bool {
	return errors.Is(e.Sentinel, target)
}

// NewDomainError creates a new domain error wrapping a sentinel error.
func NewDomainError(sentinel error) *DomainError {
	return &DomainError{Sentinel: sentinel}
}
