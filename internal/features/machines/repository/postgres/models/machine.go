// Package models defines GORM persistence models that mirror the database
// schema owned exclusively by golang-migrate. Domain code maps to and from
// these models; it never reads or writes them directly.
package models

import (
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// MachineModel is the GORM persistence model for the "machines" table.
//
// Schema is owned by golang-migrate; this struct's tags only declare how
// GORM should map Go fields to existing columns. AutoMigrate is never used.
type MachineModel struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Label     string          `gorm:"type:text;not null"`
	CoinBank  domain.CoinBank `gorm:"type:jsonb;serializer:json;not null"`
	CreatedAt time.Time       `gorm:"type:timestamptz;not null"`
	UpdatedAt time.Time       `gorm:"type:timestamptz;not null"`
}

// TableName returns the database table name for the MachineModel.
func (MachineModel) TableName() string {
	return "machines"
}

// ToDomain maps the persistence model to the domain entity.
func (m MachineModel) ToDomain() *domain.Machine {
	return &domain.Machine{
		ID:        m.ID,
		Label:     m.Label,
		CoinBank:  m.CoinBank,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// MachineModelFromDomain maps a domain entity to its persistence model.
func MachineModelFromDomain(m *domain.Machine) MachineModel {
	return MachineModel{
		ID:        m.ID,
		Label:     m.Label,
		CoinBank:  m.CoinBank,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
