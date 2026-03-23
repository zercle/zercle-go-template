package usecase

import (
	"context"
	"errors"

	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// Get implements task.Usecase.Get.
func (s *service) Get(ctx context.Context, id string) (*task.TaskResponse, error) {
	if id == "" {
		return nil, task.ErrTaskNotFound
	}

	taskID := task.TaskID(id)
	found, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, task.ErrTaskNotFound) {
			return nil, err
		}
		return nil, err
	}

	return domainToResponse(found), nil
}
