//go:build unit

package apiv1_test

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

func TestCreateMachineRequest_StructuralValidation(t *testing.T) {
	t.Parallel()

	v := validator.New()

	assert.NoError(t, v.Struct(apiv1.CreateMachineRequest{Label: "lobby-1"}))
	assert.NoError(t, v.Struct(apiv1.CreateMachineRequest{Label: "lobby-1", InitialCoins: []int32{25, 100}}))

	assert.Error(t, v.Struct(apiv1.CreateMachineRequest{Label: ""}))
	assert.Error(t, v.Struct(apiv1.CreateMachineRequest{Label: "lobby-1", InitialCoins: []int32{0}}))
	assert.Error(t, v.Struct(apiv1.CreateMachineRequest{Label: "lobby-1", InitialCoins: []int32{-5}}))
}

func TestRestockBankRequest_StructuralValidation(t *testing.T) {
	t.Parallel()

	v := validator.New()

	assert.NoError(t, v.Struct(apiv1.RestockBankRequest{Coins: []int32{25, 25, 50}}))

	assert.Error(t, v.Struct(apiv1.RestockBankRequest{}))
	assert.Error(t, v.Struct(apiv1.RestockBankRequest{Coins: []int32{}}))
	assert.Error(t, v.Struct(apiv1.RestockBankRequest{Coins: []int32{0}}))
	assert.Error(t, v.Struct(apiv1.RestockBankRequest{Coins: []int32{-5}}))
}

func TestListMachinesRequest_StructuralValidation(t *testing.T) {
	t.Parallel()

	v := validator.New()

	assert.NoError(t, v.Struct(apiv1.ListMachinesRequest{Limit: 10, Offset: 0}))
	assert.NoError(t, v.Struct(apiv1.ListMachinesRequest{}))

	assert.Error(t, v.Struct(apiv1.ListMachinesRequest{Limit: -1}))
	assert.Error(t, v.Struct(apiv1.ListMachinesRequest{Offset: -1}))
}

func TestMachinesResponseJSONFieldNames(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(apiv1.MachineResponse{ID: "m1", CoinBank: map[int32]int32{5: 2}})
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"coin_bank"`)
	assert.Contains(t, string(raw), `"created_at"`)
	assert.Contains(t, string(raw), `"updated_at"`)

	raw, err = json.Marshal(apiv1.ListMachinesResponse{})
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"machines"`)
}
