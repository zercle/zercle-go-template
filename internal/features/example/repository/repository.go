// STUB FEATURE — delete internal/features/example to start your project.

// Package repository declares the example feature's outbound (driven)
// interface. The usecase layer consumes it; the gorm implementation under
// repository/postgres satisfies it structurally.
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
)

// Repository is the outbound interface for Item persistence.
//
//go:generate go tool mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, item *domain.Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Item, error)
}
