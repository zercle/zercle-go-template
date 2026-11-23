package usecase

import (
	"context"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// List returns a paginated list of machines. It enforces safe defaults so a
// zero-value limit (e.g. no query parameter) never produces LIMIT 0 and a
// negative offset never reaches the query.
func (u *Usecase) List(ctx context.Context, limit, offset int32) ([]domain.Machine, error) {
	if limit <= 0 {
		limit = u.defaultPageSize
	}
	if limit > u.maxPageSize {
		limit = u.maxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	machines, err := u.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list machines: %w", err)
	}

	return machines, nil
}
