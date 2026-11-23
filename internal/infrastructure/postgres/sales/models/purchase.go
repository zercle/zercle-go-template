// Package models defines GORM persistence models that mirror the database
// schema owned exclusively by golang-migrate. Domain code maps to and from
// these models; it never reads or writes them directly.
package models

import (
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
)

// PurchaseModel is the GORM persistence model for the "sales_purchases" table.
// ChangeCoins is stored as a JSON array; the "serializer:json" tag makes GORM
// marshal it into and scan it back out of the jsonb column, so the domain never
// carries a driver.Valuer.
type PurchaseModel struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	MachineID          uuid.UUID `gorm:"type:uuid;not null"`
	ProductID          uuid.UUID `gorm:"type:uuid;not null"`
	PriceCents         int32     `gorm:"type:integer;not null"`
	TotalInsertedCents int32     `gorm:"type:integer;not null"`
	ChangeCents        int32     `gorm:"type:integer;not null"`
	ChangeCoins        []int32   `gorm:"type:jsonb;serializer:json;not null"`
	PurchasedAt        time.Time `gorm:"type:timestamptz;not null"`
}

// TableName returns the database table name for the PurchaseModel.
func (PurchaseModel) TableName() string {
	return "sales_purchases"
}

// ToDomain maps the persistence model to the domain record.
func (m PurchaseModel) ToDomain() *domain.PurchaseRecord {
	return &domain.PurchaseRecord{
		ID:                 m.ID,
		MachineID:          m.MachineID,
		ProductID:          m.ProductID,
		PriceCents:         m.PriceCents,
		TotalInsertedCents: m.TotalInsertedCents,
		ChangeCents:        m.ChangeCents,
		ChangeCoins:        m.ChangeCoins,
		PurchasedAt:        m.PurchasedAt,
	}
}

// PurchaseModelFromDomain maps a domain record to its persistence model.
func PurchaseModelFromDomain(r *domain.PurchaseRecord) PurchaseModel {
	return PurchaseModel{
		ID:                 r.ID,
		MachineID:          r.MachineID,
		ProductID:          r.ProductID,
		PriceCents:         r.PriceCents,
		TotalInsertedCents: r.TotalInsertedCents,
		ChangeCents:        r.ChangeCents,
		ChangeCoins:        r.ChangeCoins,
		PurchasedAt:        r.PurchasedAt,
	}
}
