package task

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// Usecase defines the contract for task business operations.
// This interface is used by handlers to invoke business logic,
// following the hexagonal architecture pattern where the application
// layer defines ports (interfaces) that adapters (handlers) use.
type Usecase interface {
	// Create creates a new task with the given parameters.
	// Returns the created task response or an error if creation fails.
	Create(ctx context.Context, input *CreateTaskInput) (*TaskResponse, error)

	// Get retrieves a task by its ID.
	// Returns the task response or ErrTaskNotFound if not found.
	Get(ctx context.Context, id string) (*TaskResponse, error)

	// List returns paginated tasks with optional filtering.
	List(ctx context.Context, params *ListParamsDTO) (*TaskListResponse, error)

	// Update modifies an existing task.
	// Returns the updated task response or an error if update fails.
	// Returns ErrTaskNotFound if the task does not exist.
	Update(ctx context.Context, id string, input *UpdateTaskInput) (*TaskResponse, error)

	// Delete removes a task.
	// Returns ErrTaskNotFound if the task does not exist.
	Delete(ctx context.Context, id string) error
}

// CreateTaskInput contains data for creating a new task.
type CreateTaskInput struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	Priority    string `json:"priority" validate:"required,oneof=low medium high"`
	UserID      string `json:"user_id" validate:"required,uuid"`
}

// UpdateTaskInput contains optional fields for updating a task.
type UpdateTaskInput struct {
	Title       *string `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	Status      *string `json:"status" validate:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority    *string `json:"priority" validate:"omitempty,oneof=low medium high"`
}

// ListParamsDTO contains query parameters for listing tasks.
type ListParamsDTO struct {
	UserID   string `query:"user_id" validate:"omitempty,uuid"`
	Status   string `query:"status" validate:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority string `query:"priority" validate:"omitempty,oneof=low medium high"`
	Limit    int32  `query:"limit" validate:"min=1,max=100"`
	Offset   int32  `query:"offset" validate:"min=0"`
}
