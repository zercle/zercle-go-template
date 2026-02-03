package task

import "time"

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

// Response represents a task in API responses.
type Response struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListResponse contains a list of tasks for API responses.
type ListResponse struct {
	Tasks  []*Response `json:"tasks"`
	Total  int64       `json:"total"`
	Limit  int32       `json:"limit"`
	Offset int32       `json:"offset"`
}

// ListQuery contains query parameters for listing tasks.
type ListQuery struct {
	UserID   string `query:"user_id" validate:"omitempty,uuid"`
	Status   string `query:"status" validate:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority string `query:"priority" validate:"omitempty,oneof=low medium high"`
	Limit    int32  `query:"limit" validate:"min=1,max=100"`
	Offset   int32  `query:"offset" validate:"min=0"`
}

// ListParams contains filtering and pagination options for listing tasks.
type ListParams struct {
	Filter Filter
	Limit  int32
	Offset int32
}

// ListResult contains paginated task results.
type ListResult struct {
	Tasks  []*Task
	Total  int64
	Limit  int32
	Offset int32
}

// Filter contains filtering options for listing tasks.
type Filter struct {
	UserID   string
	Status   *Status
	Priority *Priority
}
