// Package models defines GORM persistence models that mirror the database
// schema owned exclusively by golang-migrate. Domain code maps to and from
// these models; it never reads or writes them directly.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Item is the GORM persistence model for the "items" table.
//
// Schema is owned by golang-migrate; this struct's tags only declare how
// GORM should map Go fields to existing columns. AutoMigrate is never used.
type Item struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null"`
}

// TableName returns the database table name for the Item model.
// Declared explicitly for determinism; it also bypasses GORM's default
// pluralization rules.
func (Item) TableName() string {
	return "items"
}
