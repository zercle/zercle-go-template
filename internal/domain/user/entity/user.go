package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	FullName  string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
