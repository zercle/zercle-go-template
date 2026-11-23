package models

import (
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
)

// The two structs below are read/write projections over tables owned by other
// features (catalog_products and machines). This is a deliberate
// single-database compromise: orchestrating a purchase needs to decrement stock
// and replace the coin bank in the same transaction as the purchase insert, so
// the sales repository reaches those tables directly through its own port
// rather than calling the other features' usecases. In a truly distributed
// deployment this becomes cross-service calls or a saga, and this port is
// already the seam where that swap happens. The sales feature deliberately
// declares its own row shapes instead of importing the catalog or machines
// packages, so no feature depends on another feature's persistence internals.

// ProductRefModel projects the catalog_products columns a purchase reads and
// decrements.
type ProductRefModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	PriceCents int32     `gorm:"type:integer;not null"`
	Stock      int32     `gorm:"type:integer;not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;not null"`
}

// TableName returns the database table name for the ProductRefModel.
func (ProductRefModel) TableName() string {
	return "catalog_products"
}

// MachineRefModel projects the machines columns a purchase reads and updates.
// CoinBank uses the sales feature's own CoinBank type; the jsonb serializer
// handles marshalling both to and from the shared column.
type MachineRefModel struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey"`
	CoinBank  domain.CoinBank `gorm:"type:jsonb;serializer:json;not null"`
	CreatedAt time.Time       `gorm:"type:timestamptz;not null"`
	UpdatedAt time.Time       `gorm:"type:timestamptz;not null"`
}

// TableName returns the database table name for the MachineRefModel.
func (MachineRefModel) TableName() string {
	return "machines"
}
