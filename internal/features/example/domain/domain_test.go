//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
)

func TestSentinelErrors(t *testing.T) {
	assert.ErrorIs(t, domain.ErrItemNotFound, domain.ErrItemNotFound)
	assert.ErrorIs(t, domain.ErrInvalidName, domain.ErrInvalidName)
}
