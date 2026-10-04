// Package repository declares the catalog feature's outbound (driven)
// interface. The usecase layer consumes it; the gorm implementation under
// repository/postgres satisfies it structurally.
package repository

import (
	"context"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// Repository is the outbound interface for catalog persistence.
//
//go:generate go tool mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Product, error)
}
