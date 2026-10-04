//go:build unit

package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/sales/contract"
)

func TestPurchaseRequest_Validation(t *testing.T) {
	t.Parallel()

	v := validator.New()

	assert.NoError(t, v.Struct(contract.PurchaseRequest{
		MachineID: "12345678-1234-1234-1234-123456789abc",
		ProductID: "22345678-1234-1234-1234-123456789abc",
		Coins:     []int32{25, 25, 50},
	}))

	assert.Error(t, v.Struct(contract.PurchaseRequest{MachineID: "", ProductID: "22345678-1234-1234-1234-123456789abc", Coins: []int32{5}}))
	assert.Error(t, v.Struct(contract.PurchaseRequest{MachineID: "not-a-uuid", ProductID: "22345678-1234-1234-1234-123456789abc", Coins: []int32{5}}))
	assert.Error(t, v.Struct(contract.PurchaseRequest{MachineID: "12345678-1234-1234-1234-123456789abc", ProductID: "", Coins: []int32{5}}))
	assert.Error(t, v.Struct(contract.PurchaseRequest{MachineID: "12345678-1234-1234-1234-123456789abc", ProductID: "not-a-uuid", Coins: []int32{5}}))
	assert.Error(t, v.Struct(contract.PurchaseRequest{MachineID: "12345678-1234-1234-1234-123456789abc", ProductID: "22345678-1234-1234-1234-123456789abc"}))
	assert.Error(t, v.Struct(contract.PurchaseRequest{MachineID: "12345678-1234-1234-1234-123456789abc", ProductID: "22345678-1234-1234-1234-123456789abc", Coins: []int32{0}}))
	assert.Error(t, v.Struct(contract.PurchaseRequest{MachineID: "12345678-1234-1234-1234-123456789abc", ProductID: "22345678-1234-1234-1234-123456789abc", Coins: []int32{-5}}))
}

func TestResponseJSONFieldNames(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(contract.PurchaseResponse{ID: "s1"})
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"machine_id"`)
	assert.Contains(t, string(raw), `"product_id"`)
	assert.Contains(t, string(raw), `"total_inserted_cents"`)
	assert.Contains(t, string(raw), `"change_coins"`)
	assert.Contains(t, string(raw), `"purchased_at"`)
}
