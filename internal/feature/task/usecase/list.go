package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// List implements task.Usecase.List.
func (s *service) List(ctx context.Context, params *task.ListParamsDTO) (*task.TaskListResponse, error) {
	if params == nil {
		params = &task.ListParamsDTO{}
	}

	// Set default pagination if not provided
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	// Convert params to repository params
	repoParams := listParamsToRepo(*params)

	// Call repository
	result, err := s.repo.List(ctx, &repoParams)
	if err != nil {
		return nil, err
	}

	// Convert result to DTO
	return listResultToDTO(result), nil
}
