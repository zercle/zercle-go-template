// Package port declares the machines feature's outbound (driven) interface.
// The usecase layer consumes it; the gorm implementation under
// internal/infrastructure/postgres/machines satisfies it structurally.
package port

import (
	"context"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// Repository is the outbound interface for machine persistence.
//
//go:generate go tool mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, machine *domain.Machine) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Machine, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Machine, error)
	// RestockBank adds coins to the machine's coin bank inside one transaction,
	// locking the row so concurrent restocks cannot lose an update. A missing
	// machine maps to domain.ErrMachineNotFound.
	RestockBank(ctx context.Context, id uuid.UUID, coins []int32) error
}
