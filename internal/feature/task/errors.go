package task

import "errors"

var (
	// ErrTaskNotFound is returned when a task is not found.
	ErrTaskNotFound = errors.New("task not found")
	// ErrInvalidTask is returned when task data is invalid.
	ErrInvalidTask = errors.New("invalid task")
)
