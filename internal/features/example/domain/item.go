// STUB FEATURE — delete internal/features/example to start your project.

package domain

import (
	"time"

	"github.com/google/uuid"
)

// Item is the trivial example entity.
type Item struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Rename updates the item name and refreshes the updated-at timestamp.
func (i *Item) Rename(name string) {
	i.Name = name
	i.UpdatedAt = time.Now().UTC()
}
