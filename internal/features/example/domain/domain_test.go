//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
)

func TestItem_Rename(t *testing.T) {
	item := &domain.Item{
		ID:        uuid.New(),
		Name:      "original",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC().Add(-1 * time.Hour),
	}

	before := item.UpdatedAt
	item.Rename("renamed")

	assert.Equal(t, "renamed", item.Name)
	assert.True(t, item.UpdatedAt.After(before), "expected UpdatedAt to advance")
}

func TestSentinelErrors(t *testing.T) {
	assert.ErrorIs(t, domain.ErrItemNotFound, domain.ErrItemNotFound)
	assert.ErrorIs(t, domain.ErrInvalidName, domain.ErrInvalidName)
}
