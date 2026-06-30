//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package dto_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/zercle/zercle-go-template/internal/features/example/dto"
)

func TestCreateItemRequest_Validation(t *testing.T) {
	v := validator.New()

	valid := dto.CreateItemRequest{Name: "valid name"}
	assert.NoError(t, v.Struct(valid))

	empty := dto.CreateItemRequest{Name: ""}
	assert.Error(t, v.Struct(empty))

	long := dto.CreateItemRequest{Name: string(make([]byte, 256))}
	assert.Error(t, v.Struct(long))
}

func TestListItemsRequest_Validation(t *testing.T) {
	v := validator.New()

	valid := dto.ListItemsRequest{Limit: 10, Offset: 0}
	assert.NoError(t, v.Struct(valid))

	defaultLimit := dto.ListItemsRequest{}
	assert.NoError(t, v.Struct(defaultLimit))

	highLimit := dto.ListItemsRequest{Limit: 101, Offset: 0}
	assert.Error(t, v.Struct(highLimit))

	negativeOffset := dto.ListItemsRequest{Limit: 10, Offset: -1}
	assert.Error(t, v.Struct(negativeOffset))
}
