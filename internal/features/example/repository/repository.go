// STUB FEATURE — delete internal/features/example to start your project.

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	sqlcdb "github.com/zercle/zercle-go-template/internal/infrastructure/db/sqlc"
)

// Repository is a pgx + sqlc implementation of the domain.Repository port.
type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlcdb.Queries
}

// NewRepository returns a Repository backed by the provided pool and sqlc queries.
func NewRepository(pool *pgxpool.Pool, queries *sqlcdb.Queries) *Repository {
	return &Repository{pool: pool, queries: queries}
}

// Create persists a new item.
// nolint:wrapcheck // sqlc exec error is propagated without added context.
func (r *Repository) Create(ctx context.Context, item *domain.Item) error {
	return r.queries.CreateItem(ctx, sqlcdb.CreateItemParams{
		ID:        item.ID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	})
}

// GetByID retrieves an item by its UUID. It maps pgx.ErrNoRows to
// domain.ErrItemNotFound and wraps other errors.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	row, err := r.queries.GetItem(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}

	return mapRowToDomain(row), nil
}

// List returns a paginated slice of items ordered by created_at descending.
func (r *Repository) List(ctx context.Context, limit, offset int32) ([]domain.Item, error) {
	rows, err := r.queries.ListItems(ctx, sqlcdb.ListItemsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	items := make([]domain.Item, len(rows))
	for i, row := range rows {
		items[i] = *mapRowToDomain(row)
	}

	return items, nil
}

func mapRowToDomain(row sqlcdb.Item) *domain.Item {
	return &domain.Item{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
