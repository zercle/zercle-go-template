// Package task provides domain entities and contracts for the task feature.
// This package follows clean/hexagonal architecture principles where the domain
// layer defines contracts (interfaces) that infrastructure components implement.
package task

import (
	"errors"
	"time"

	"github.com/zercle/zercle-go-template/pkg/uid"
)

// ID is a typed identifier for Task entities.
type ID string

// Status represents the current state of a task.
type Status string

const (
	// StatusPending indicates a task that has not been started yet.
	StatusPending Status = "pending"
	// StatusInProgress indicates a task that is currently being worked on.
	StatusInProgress Status = "in_progress"
	// StatusCompleted indicates a task that has been finished.
	StatusCompleted Status = "completed"
	// StatusCancelled indicates a task that has been cancelled.
	StatusCancelled Status = "cancelled"
)

// IsValid checks if the status is a valid value.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

// Priority represents the priority level of a task.
type Priority string

const (
	// PriorityLow indicates a low priority task.
	PriorityLow Priority = "low"
	// PriorityMedium indicates a medium priority task.
	PriorityMedium Priority = "medium"
	// PriorityHigh indicates a high priority task.
	PriorityHigh Priority = "high"
)

// IsValid checks if the priority is a valid value.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// Task is the domain entity representing a task.
type Task struct {
	ID          ID
	Title       string
	Description string
	Status      Status
	Priority    Priority
	UserID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// New creates a new Task with generated ID and timestamps.
func New(title, description string, priority Priority, userID string) (*Task, error) {
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
		ID:          ID(uid.New().String()),
		Title:       title,
		Description: description,
		Status:      StatusPending,
		Priority:    priority,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewWithID creates a Task with specified values (for reconstruction from database).
func NewWithID(id ID, title, description string, status Status, priority Priority, userID string, createdAt, updatedAt time.Time) (*Task, error) {
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
	t.Status = StatusInProgress
	t.UpdatedAt = time.Now().UTC()
}

// MarkCompleted sets task status to completed.
func (t *Task) MarkCompleted() {
	t.Status = StatusCompleted
	t.UpdatedAt = time.Now().UTC()
}

// Cancel sets task status to cancelled.
func (t *Task) Cancel() {
	t.Status = StatusCancelled
	t.UpdatedAt = time.Now().UTC()
}
