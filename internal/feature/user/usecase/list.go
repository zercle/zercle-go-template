package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// List implements user.Usecase.List.
func (s *service) List(ctx context.Context, params *user.ListParamsDTO) (*user.ListResultDTO, error) {
	if params == nil {
		params = &user.ListParamsDTO{}
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
		s.logger.Error().Err(err).Msg("failed to list users")
		return nil, err
	}

	// Convert result to DTO
	return listResultToDTO(result), nil
}
