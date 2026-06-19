// STUB FEATURE — delete internal/features/example to start your project.

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/models"
)

// Repository is a GORM implementation of the domain.Repository port.
type Repository struct {
	db *gorm.DB
}

// NewRepository returns a Repository backed by the provided *gorm.DB.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTx returns a Repository that uses the provided transactional *gorm.DB.
// It mirrors the prior sqlc Queries.WithTx API and is intended for callers
// that need to compose multiple repository calls inside a single transaction.
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// Create persists a new item.
func (r *Repository) Create(ctx context.Context, item *domain.Item) error {
	m := mapDomainToModel(item)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("create item: %w", err)
	}
	return nil
}

// GetByID retrieves an item by its UUID. It maps gorm.ErrRecordNotFound to
// domain.ErrItemNotFound via errors.Is and wraps other errors.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	var m models.Item
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	return mapModelToDomain(&m), nil
}

// List returns a paginated slice of items ordered by created_at descending,
// then by id descending to keep order stable across pages with identical
// timestamps.
func (r *Repository) List(ctx context.Context, limit, offset int32) ([]domain.Item, error) {
	var ms []models.Item
	if err := r.db.WithContext(ctx).
		Order("created_at DESC, id DESC").
		Limit(int(limit)).
		Offset(int(offset)).
		Find(&ms).Error; err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	items := make([]domain.Item, len(ms))
	for i := range ms {
		items[i] = *mapModelToDomain(&ms[i])
	}
	return items, nil
}

func mapModelToDomain(m *models.Item) *domain.Item {
	return &domain.Item{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func mapDomainToModel(item *domain.Item) models.Item {
	return models.Item{
		ID:        item.ID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
