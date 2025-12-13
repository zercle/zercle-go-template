package domain

import (
	"time"

	"github.com/google/uuid"
	sharedDomain "github.com/zercle/zercle-go-template/internal/shared/domain"
)

// Post represents a blog post (Domain Entity).
type Post struct {
	ID        uuid.UUID
	Title     string
	Content   string
	AuthorID  sharedDomain.UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}
