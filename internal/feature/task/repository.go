package task

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// ListParams contains filtering and pagination options for listing tasks.
type ListParams struct {
	// Filter contains the filtering criteria.
	Filter TaskFilter
	// Limit is the maximum number of results to return.
	Limit int32
	// Offset is the number of results to skip.
	Offset int32
}

// ListResult contains paginated task results.
type ListResult struct {
	// Tasks is the list of tasks matching the query.
	Tasks []*Task
	// Total is the total count of matching records (before pagination).
	Total int64
	// Limit is the limit used for this query.
	Limit int32
	// Offset is the offset used for this query.
	Offset int32
}

// Repository defines the contract for task data persistence.
// This interface follows the repository pattern for clean architecture,
// allowing the domain layer to define data access contracts that
// infrastructure components implement.
type Repository interface {
	// Create inserts a new task into the data store.
	// Returns the created task with generated ID and timestamps.
	Create(ctx context.Context, task *Task) (*Task, error)

	// GetByID retrieves a task by its unique identifier.
	// Returns ErrTaskNotFound if no task exists with the given ID.
	GetByID(ctx context.Context, id TaskID) (*Task, error)

	// Update modifies an existing task in the data store.
	// Returns the updated task.
	// Returns ErrTaskNotFound if no task exists with the given ID.
	Update(ctx context.Context, task *Task) (*Task, error)

	// Delete removes a task from the data store.
	// Returns ErrTaskNotFound if no task exists with the given ID.
	Delete(ctx context.Context, id TaskID) error

	// List returns a paginated list of tasks matching the given parameters.
	List(ctx context.Context, params *ListParams) (*ListResult, error)

	// Count returns the total number of tasks matching the given filter.
	Count(ctx context.Context, filter TaskFilter) (int64, error)

	// ExistsByID checks if a task with the given ID exists.
	ExistsByID(ctx context.Context, id TaskID) (bool, error)
}
