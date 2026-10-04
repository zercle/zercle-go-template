// STUB FEATURE — delete internal/features/example to start your project.

package domain

import (
	"time"
	"uuid"
)

// Item is the trivial example entity.
type Item struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
