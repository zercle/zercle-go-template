package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// Update implements task.Usecase.Update.
func (s *service) Update(ctx context.Context, id string, input *task.UpdateTaskInput) (*task.TaskResponse, error) {
	if id == "" {
		return nil, task.ErrTaskNotFound
	}
	if input == nil {
		return nil, errors.New("update input is required")
	}

	taskID := task.TaskID(id)

	// Fetch existing entity
	existing, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, task.ErrTaskNotFound) {
			return nil, err
		}
		return nil, err
	}

	// Validate status if being changed
	if input.Status != nil {
		if !task.TaskStatus(*input.Status).IsValid() {
			return nil, task.ErrInvalidTaskStatus
		}
	}

	// Validate priority if being changed
	if input.Priority != nil {
		if !task.TaskPriority(*input.Priority).IsValid() {
			return nil, task.ErrInvalidTaskPriority
		}
	}

	// Apply updates
	updated := &task.Task{
		ID:          existing.ID,
		Title:       existing.Title,
		Description: existing.Description,
		Status:      existing.Status,
		Priority:    existing.Priority,
		UserID:      existing.UserID,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
	}

	if input.Title != nil {
		updated.Title = *input.Title
	}
	if input.Description != nil {
		updated.Description = *input.Description
	}
	if input.Status != nil {
		updated.Status = task.TaskStatus(*input.Status)
	}
	if input.Priority != nil {
		updated.Priority = task.TaskPriority(*input.Priority)
	}

	// Persist changes
	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		if errors.Is(err, task.ErrTaskNotFound) {
			return nil, err
		}
		return nil, err
	}

	return domainToResponse(result), nil
}
