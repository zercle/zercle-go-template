package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/repository/postgres/models"
)

// Repository is a GORM implementation of the repository.Repository interface.
type Repository struct {
	db *gorm.DB
}

// NewRepository returns a Repository backed by the provided *gorm.DB.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create persists a new machine.
func (r *Repository) Create(ctx context.Context, machine *domain.Machine) error {
	if machine == nil {
		return fmt.Errorf("create machine: nil machine")
	}
	m := models.MachineModelFromDomain(machine)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("create machine: %w", err)
	}
	return nil
}

// GetByID retrieves a machine by its UUID. It maps gorm.ErrRecordNotFound to
// domain.ErrMachineNotFound via errors.Is and wraps other errors.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Machine, error) {
	var m models.MachineModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMachineNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get machine: %w", err)
	}
	return m.ToDomain(), nil
}

// List returns a paginated slice of machines ordered by created_at descending,
// then by id descending to keep order stable across pages with identical
// timestamps.
func (r *Repository) List(ctx context.Context, limit, offset int32) ([]domain.Machine, error) {
	var ms []models.MachineModel
	if err := r.db.WithContext(ctx).
		Order("created_at DESC, id DESC").
		Limit(int(limit)).
		Offset(int(offset)).
		Find(&ms).Error; err != nil {
		return nil, fmt.Errorf("list machines: %w", err)
	}

	machines := make([]domain.Machine, len(ms))
	for i := range ms {
		machines[i] = *ms[i].ToDomain()
	}
	return machines, nil
}

// RestockBank adds coins to the machine's coin bank inside a transaction. The
// machine row is locked with SELECT ... FOR UPDATE so two concurrent restocks
// cannot lose an update, the new bank is composed with domain.AddCoins, and the
// coin_bank column is written back. A missing machine maps to
// domain.ErrMachineNotFound.
func (r *Repository) RestockBank(ctx context.Context, id uuid.UUID, coins []int32) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m models.MachineModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&m, "id = ?", id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrMachineNotFound
		}
		if err != nil {
			return fmt.Errorf("lock machine: %w", err)
		}

		bankAfter := domain.AddCoins(m.CoinBank, coins)
		update := models.MachineModel{CoinBank: bankAfter, UpdatedAt: time.Now().UTC()}
		if err := tx.Model(&models.MachineModel{}).
			Where("id = ?", id).
			Select("coin_bank").
			Updates(&update).Error; err != nil {
			return fmt.Errorf("update coin bank: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("restock bank: %w", err)
	}
	return nil
}
