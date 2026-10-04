//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package contract_test

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/zercle/zercle-go-template/internal/features/example/contract"
)

// TestCreateItemRequest_StructuralValidation pins that the wire type only
// enforces structural constraints. The length cap is deployment-configurable
// and enforced in the usecase layer, so a name longer than any default must
// pass the tag: otherwise the tag would silently cap the configurable limit.
func TestCreateItemRequest_StructuralValidation(t *testing.T) {
	v := validator.New()

	assert.NoError(t, v.Struct(contract.CreateItemRequest{Name: "valid name"}))

	// Far beyond every default limit but still structurally valid.
	assert.NoError(t, v.Struct(contract.CreateItemRequest{Name: strings.Repeat("a", 5000)}))

	// Empty name is structurally invalid regardless of configuration.
	assert.Error(t, v.Struct(contract.CreateItemRequest{Name: ""}))
}

// TestListItemsRequest_StructuralValidation pins that pagination bounds are not
// hardcoded in the wire type: any non-negative limit passes the tag and the
// usecase layer clamps it to the configured maximum.
func TestListItemsRequest_StructuralValidation(t *testing.T) {
	v := validator.New()

	assert.NoError(t, v.Struct(contract.ListItemsRequest{Limit: 10, Offset: 0}))
	assert.NoError(t, v.Struct(contract.ListItemsRequest{}))
	assert.NoError(t, v.Struct(contract.ListItemsRequest{Limit: 100_000, Offset: 0}))

	assert.Error(t, v.Struct(contract.ListItemsRequest{Limit: -1}))
	assert.Error(t, v.Struct(contract.ListItemsRequest{Offset: -1}))
}
