// Package fixtures provides sample domain objects for tests.
package fixtures

import (
	"time"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
)

// NewItem returns a sample Item with the given name. It uses a deterministic
// generated UUID for the ID and fixed timestamps so tests can assert against
// known values.
func NewItem(name string) domain.Item {
	return domain.Item{
		ID:        uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
		Name:      name,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}
