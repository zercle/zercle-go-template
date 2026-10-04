package postgres

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres/models"
)

// Repository is a GORM implementation of the repository.Repository interface.
type Repository struct {
	db *gorm.DB
}

// NewRepository returns a Repository backed by the provided *gorm.DB.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create persists a new product.
func (r *Repository) Create(ctx context.Context, product *domain.Product) error {
	if product == nil {
		return fmt.Errorf("create product: nil product")
	}
	m := models.ProductModelFromDomain(product)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

// GetByID retrieves a product by its UUID. It maps gorm.ErrRecordNotFound to
// domain.ErrProductNotFound via errors.Is and wraps other errors.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var m models.ProductModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	return m.ToDomain(), nil
}

// List returns a paginated slice of products ordered by created_at descending,
// then by id descending to keep order stable across pages with identical
// timestamps.
func (r *Repository) List(ctx context.Context, limit, offset int32) ([]domain.Product, error) {
	var ms []models.ProductModel
	if err := r.db.WithContext(ctx).
		Order("created_at DESC, id DESC").
		Limit(int(limit)).
		Offset(int(offset)).
		Find(&ms).Error; err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	products := make([]domain.Product, len(ms))
	for i := range ms {
		products[i] = *ms[i].ToDomain()
	}
	return products, nil
}
