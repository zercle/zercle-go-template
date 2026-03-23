package usecase

import (
	"context"
	"errors"

	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// Delete implements task.Usecase.Delete.
func (s *service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return task.ErrTaskNotFound
	}

	taskID := task.TaskID(id)

	// Check if task exists first
	exists, err := s.repo.ExistsByID(ctx, taskID)
	if err != nil {
		return err
	}
	if !exists {
		return task.ErrTaskNotFound
	}

	// Delete via repository
	if err := s.repo.Delete(ctx, taskID); err != nil {
		if errors.Is(err, task.ErrTaskNotFound) {
			return err
		}
		return err
	}

	return nil
}
