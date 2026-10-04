//go:build unit

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	assert.ErrorIs(t, domain.ErrProductNotFound, domain.ErrProductNotFound)
	assert.ErrorIs(t, domain.ErrInvalidID, domain.ErrInvalidID)
	assert.ErrorIs(t, domain.ErrInvalidProductName, domain.ErrInvalidProductName)
	assert.ErrorIs(t, domain.ErrInvalidPrice, domain.ErrInvalidPrice)
}
