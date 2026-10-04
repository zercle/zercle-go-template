//go:build unit

package contract_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
)

// TestCreateProductRequest_StructuralValidation pins that the wire type only
// enforces structural constraints. The name-length cap is deployment-
// configurable and enforced in the usecase layer, so a long name must pass
// the tag: otherwise the tag would silently cap the configurable limit.
func TestCreateProductRequest_StructuralValidation(t *testing.T) {
	t.Parallel()

	v := validator.New()

	assert.NoError(t, v.Struct(contract.CreateProductRequest{Name: "Cola", PriceCents: 150, Stock: 3}))
	assert.NoError(t, v.Struct(contract.CreateProductRequest{Name: strings.Repeat("a", 5000), PriceCents: 150, Stock: 0}))

	assert.Error(t, v.Struct(contract.CreateProductRequest{Name: "", PriceCents: 150, Stock: 3}))
	assert.Error(t, v.Struct(contract.CreateProductRequest{Name: "Cola", PriceCents: 0, Stock: 3}))
	assert.Error(t, v.Struct(contract.CreateProductRequest{Name: "Cola", PriceCents: -1, Stock: 3}))
	assert.Error(t, v.Struct(contract.CreateProductRequest{Name: "Cola", PriceCents: 150, Stock: -1}))
}

// TestListProductsRequest_StructuralValidation pins that pagination bounds are
// not hardcoded in the wire type: any non-negative limit passes the tag and
// the usecase layer clamps it to the configured maximum.
func TestListProductsRequest_StructuralValidation(t *testing.T) {
	t.Parallel()

	v := validator.New()

	assert.NoError(t, v.Struct(contract.ListProductsRequest{Limit: 10, Offset: 0}))
	assert.NoError(t, v.Struct(contract.ListProductsRequest{}))
	assert.NoError(t, v.Struct(contract.ListProductsRequest{Limit: 100_000, Offset: 0}))

	assert.Error(t, v.Struct(contract.ListProductsRequest{Limit: -1}))
	assert.Error(t, v.Struct(contract.ListProductsRequest{Offset: -1}))
}

func TestResponseJSONFieldNames(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(contract.ProductResponse{ID: "p1"})
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"price_cents"`)
	assert.Contains(t, string(raw), `"created_at"`)
	assert.Contains(t, string(raw), `"updated_at"`)

	raw, err = json.Marshal(contract.ListProductsResponse{})
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"products"`)
}
