// Package models defines GORM persistence models that mirror the database
// schema owned exclusively by golang-migrate. Domain code maps to and from
// these models; it never reads or writes them directly.
package models

import (
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// ProductModel is the GORM persistence model for the "catalog_products" table.
//
// Schema is owned by golang-migrate; this struct's tags only declare how
// GORM should map Go fields to existing columns. AutoMigrate is never used.
type ProductModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name       string    `gorm:"type:text;not null"`
	PriceCents int32     `gorm:"type:integer;not null"`
	Stock      int32     `gorm:"type:integer;not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;not null"`
}

// TableName returns the database table name for the ProductModel.
func (ProductModel) TableName() string {
	return "catalog_products"
}

// ToDomain maps the persistence model to the domain entity.
func (m ProductModel) ToDomain() *domain.Product {
	return &domain.Product{
		ID:         m.ID,
		Name:       m.Name,
		PriceCents: m.PriceCents,
		Stock:      m.Stock,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// ProductModelFromDomain maps a domain entity to its persistence model.
func ProductModelFromDomain(p *domain.Product) ProductModel {
	return ProductModel{
		ID:         p.ID,
		Name:       p.Name,
		PriceCents: p.PriceCents,
		Stock:      p.Stock,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}
