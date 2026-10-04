package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/repository/postgres/models"
)

// Repository is a GORM implementation of the repository.Repository interface.
type Repository struct {
	db *gorm.DB
}

// NewRepository returns a Repository backed by the provided *gorm.DB.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetProduct reads the catalog projection for a product. It maps
// gorm.ErrRecordNotFound to domain.ErrProductNotFound via errors.Is and wraps
// other errors.
func (r *Repository) GetProduct(ctx context.Context, productID uuid.UUID) (domain.SaleProduct, error) {
	var m models.ProductRefModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", productID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.SaleProduct{}, domain.ErrProductNotFound
	}
	if err != nil {
		return domain.SaleProduct{}, fmt.Errorf("get product: %w", err)
	}
	return domain.SaleProduct{ProductID: m.ID, PriceCents: m.PriceCents, Stock: m.Stock}, nil
}

// GetMachineBank reads a machine's current coin bank. It maps
// gorm.ErrRecordNotFound to domain.ErrMachineNotFound via errors.Is and wraps
// other errors.
func (r *Repository) GetMachineBank(ctx context.Context, machineID uuid.UUID) (domain.CoinBank, error) {
	var m models.MachineRefModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", machineID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMachineNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get machine bank: %w", err)
	}
	return m.CoinBank, nil
}

// CommitPurchase atomically records a purchase and applies its effects: the
// product's stock is decremented, the machine's coin bank is replaced with
// bankAfter, and the purchase record is inserted. The product row is locked
// with SELECT ... FOR UPDATE so two concurrent buyers of the last unit cannot
// both pass the stock check.
//
// A missing product maps to domain.ErrProductNotFound; losing the stock race
// (the guarded UPDATE affects no row) maps to domain.ErrOutOfStock; a missing
// machine maps to domain.ErrMachineNotFound. All statements share one
// transaction, so a failure anywhere rolls back the whole purchase.
func (r *Repository) CommitPurchase(ctx context.Context, machineID, productID uuid.UUID, record *domain.PurchaseRecord, bankAfter domain.CoinBank) error {
	if record == nil {
		return fmt.Errorf("commit purchase: nil record")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product models.ProductRefModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&product, "id = ?", productID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrProductNotFound
		}
		if err != nil {
			return fmt.Errorf("lock product: %w", err)
		}

		res := tx.Model(&models.ProductRefModel{}).
			Where("id = ? AND stock > 0", productID).
			Update("stock", gorm.Expr("stock - 1"))
		if res.Error != nil {
			return fmt.Errorf("decrement stock: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return domain.ErrOutOfStock
		}

		bank := models.MachineRefModel{CoinBank: bankAfter, UpdatedAt: time.Now().UTC()}
		machineRes := tx.Model(&models.MachineRefModel{}).
			Where("id = ?", machineID).
			Select("coin_bank").
			Updates(&bank)
		if machineRes.Error != nil {
			return fmt.Errorf("update machine bank: %w", machineRes.Error)
		}
		if machineRes.RowsAffected == 0 {
			return domain.ErrMachineNotFound
		}

		purchase := models.PurchaseModelFromDomain(record)
		if err := tx.Create(&purchase).Error; err != nil {
			return fmt.Errorf("insert purchase: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("commit purchase: %w", err)
	}
	return nil
}
