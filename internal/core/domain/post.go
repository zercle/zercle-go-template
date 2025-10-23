package domain

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a blog post (Domain Entity).
type Post struct {
	ID        uuid.UUID
	Title     string
	Content   string
	AuthorID  uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
