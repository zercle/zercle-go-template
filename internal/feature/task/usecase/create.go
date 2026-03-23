package usecase

import (
	"context"
	"errors"

	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// Create implements task.Usecase.Create.
func (s *service) Create(ctx context.Context, input *task.CreateTaskInput) (*task.TaskResponse, error) {
	// Validate input
	if input == nil {
		return nil, errors.New("create input is required")
	}
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	if input.Priority == "" {
		return nil, errors.New("priority is required")
	}
	if input.UserID == "" {
		return nil, errors.New("user_id is required")
	}

	// Validate priority
	if !task.TaskPriority(input.Priority).IsValid() {
		return nil, task.ErrInvalidTaskPriority
	}

	// Create domain entity
	newTask, err := createInputToDomain(input)
	if err != nil {
		return nil, err
	}

	// Persist via repository
	created, err := s.repo.Create(ctx, newTask)
	if err != nil {
		return nil, err
	}

	return domainToResponse(created), nil
}
