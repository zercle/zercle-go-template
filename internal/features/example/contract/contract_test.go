//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package contract_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/zercle/zercle-go-template/internal/features/example/contract"
)

func TestCreateItemRequest_Validation(t *testing.T) {
	v := validator.New()

	valid := contract.CreateItemRequest{Name: "valid name"}
	assert.NoError(t, v.Struct(valid))

	empty := contract.CreateItemRequest{Name: ""}
	assert.Error(t, v.Struct(empty))

	long := contract.CreateItemRequest{Name: string(make([]byte, 256))}
	assert.Error(t, v.Struct(long))
}

func TestListItemsRequest_Validation(t *testing.T) {
	v := validator.New()

	valid := contract.ListItemsRequest{Limit: 10, Offset: 0}
	assert.NoError(t, v.Struct(valid))

	defaultLimit := contract.ListItemsRequest{}
	assert.NoError(t, v.Struct(defaultLimit))

	highLimit := contract.ListItemsRequest{Limit: 101, Offset: 0}
	assert.Error(t, v.Struct(highLimit))

	negativeOffset := contract.ListItemsRequest{Limit: 10, Offset: -1}
	assert.Error(t, v.Struct(negativeOffset))
}
