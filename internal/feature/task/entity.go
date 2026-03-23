// Package task provides domain entities and contracts for the task feature.
// This package follows clean/hexagonal architecture principles where the domain
// layer defines contracts (interfaces) that infrastructure components implement.
package task

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TaskID is a typed identifier for Task entities.
type TaskID string

// TaskStatus represents the current state of a task.
type TaskStatus string

const (
	// TaskStatusPending indicates a task that has not been started yet.
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusInProgress indicates a task that is currently being worked on.
	TaskStatusInProgress TaskStatus = "in_progress"
	// TaskStatusCompleted indicates a task that has been finished.
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusCancelled indicates a task that has been cancelled.
	TaskStatusCancelled TaskStatus = "cancelled"
)

// IsValid checks if the status is a valid value.
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// TaskPriority represents the priority level of a task.
type TaskPriority string

const (
	// TaskPriorityLow indicates a low priority task.
	TaskPriorityLow TaskPriority = "low"
	// TaskPriorityMedium indicates a medium priority task.
	TaskPriorityMedium TaskPriority = "medium"
	// TaskPriorityHigh indicates a high priority task.
	TaskPriorityHigh TaskPriority = "high"
)

// IsValid checks if the priority is a valid value.
func (p TaskPriority) IsValid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh:
		return true
	default:
		return false
	}
}

// Task is the domain entity representing a task.
type Task struct {
	ID          TaskID
	Title       string
	Description string
	Status      TaskStatus
	Priority    TaskPriority
	UserID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTask creates a new Task with generated ID and timestamps.
func NewTask(title, description string, priority TaskPriority, userID string) (*Task, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if !priority.IsValid() {
		return nil, errors.New("invalid task priority")
	}

	now := time.Now().UTC()
	return &Task{
		ID:          TaskID(uuid.New().String()),
		Title:       title,
		Description: description,
		Status:      TaskStatusPending,
		Priority:    priority,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewTaskWithID creates a Task with specified values (for reconstruction from database).
func NewTaskWithID(id TaskID, title, description string, status TaskStatus, priority TaskPriority, userID string, createdAt, updatedAt time.Time) (*Task, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if title == "" {
		return nil, errors.New("title is required")
	}
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if !status.IsValid() {
		return nil, errors.New("invalid task status")
	}
	if !priority.IsValid() {
		return nil, errors.New("invalid task priority")
	}

	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      status,
		Priority:    priority,
		UserID:      userID,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// MarkInProgress sets task status to in_progress.
func (t *Task) MarkInProgress() {
	t.Status = TaskStatusInProgress
	t.UpdatedAt = time.Now().UTC()
}

// MarkCompleted sets task status to completed.
func (t *Task) MarkCompleted() {
	t.Status = TaskStatusCompleted
	t.UpdatedAt = time.Now().UTC()
}

// Cancel sets task status to cancelled.
func (t *Task) Cancel() {
	t.Status = TaskStatusCancelled
	t.UpdatedAt = time.Now().UTC()
}

// CreateTaskRequest contains data for creating a new task.
type CreateTaskRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	Priority    string `json:"priority" validate:"required,oneof=low medium high"`
	UserID      string `json:"user_id" validate:"required,uuid"`
}

// UpdateTaskRequest contains optional fields for updating a task.
type UpdateTaskRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	Status      *string `json:"status" validate:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority    *string `json:"priority" validate:"omitempty,oneof=low medium high"`
}

// TaskResponse represents a task in API responses.
type TaskResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskListResponse contains a list of tasks for API responses.
type TaskListResponse struct {
	Tasks  []*TaskResponse `json:"tasks"`
	Total  int64           `json:"total"`
	Limit  int32           `json:"limit"`
	Offset int32           `json:"offset"`
}

// TaskFilter contains filtering options for listing tasks.
type TaskFilter struct {
	UserID   string
	Status   *TaskStatus
	Priority *TaskPriority
}
