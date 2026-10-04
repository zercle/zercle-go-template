//go:build unit

package apiv1_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

func TestCatalogAliasesRoundTripJSON(t *testing.T) {
	t.Parallel()

	req := apiv1.CreateProductRequest{Name: "widget", PriceCents: 250, Stock: 3}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"widget","price_cents":250,"stock":3}`, string(data))

	resp := apiv1.ProductResponse{
		ID: "p1", Name: "widget", PriceCents: 250, Stock: 3,
		CreatedAt: "t1", UpdatedAt: "t2",
	}
	data, err = json.Marshal(apiv1.ListProductsResponse{Products: []apiv1.ProductResponse{resp}})
	require.NoError(t, err)
	require.JSONEq(t,
		`{"products":[{"id":"p1","name":"widget","price_cents":250,"stock":3,"created_at":"t1","updated_at":"t2"}]}`,
		string(data))
}

func TestMachinesAliasesRoundTripJSON(t *testing.T) {
	t.Parallel()

	req := apiv1.CreateMachineRequest{Label: "m1", InitialCoins: []int32{25, 100}}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	require.JSONEq(t, `{"label":"m1","initial_coins":[25,100]}`, string(data))

	data, err = json.Marshal(apiv1.RestockBankRequest{Coins: []int32{5, 5}})
	require.NoError(t, err)
	require.JSONEq(t, `{"coins":[5,5]}`, string(data))

	resp := apiv1.MachineResponse{
		ID: "m1", Label: "lobby", CoinBank: map[int32]int32{25: 2, 100: 1},
		CreatedAt: "t1", UpdatedAt: "t2",
	}
	data, err = json.Marshal(apiv1.ListMachinesResponse{Machines: []apiv1.MachineResponse{resp}})
	require.NoError(t, err)
	require.JSONEq(t,
		`{"machines":[{"id":"m1","label":"lobby","coin_bank":{"25":2,"100":1},"created_at":"t1","updated_at":"t2"}]}`,
		string(data))
}

func TestSalesAliasesRoundTripJSON(t *testing.T) {
	t.Parallel()

	req := apiv1.PurchaseRequest{
		MachineID: "m1", ProductID: "p1", Coins: []int32{100},
	}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	require.JSONEq(t, `{"machine_id":"m1","product_id":"p1","coins":[100]}`, string(data))

	resp := apiv1.PurchaseResponse{
		ID: "s1", MachineID: "m1", ProductID: "p1",
		PriceCents: 50, TotalInsertedCents: 100, ChangeCents: 50,
		ChangeCoins: []int32{50}, PurchasedAt: "t1",
	}
	data, err = json.Marshal(resp)
	require.NoError(t, err)
	require.JSONEq(t,
		`{"id":"s1","machine_id":"m1","product_id":"p1","price_cents":50,"total_inserted_cents":100,"change_cents":50,"change_coins":[50],"purchased_at":"t1"}`,
		string(data))
}

func TestErrCodeReExports(t *testing.T) {
	t.Parallel()

	require.Equal(t, "NOT_FOUND", apiv1.ErrCodeNotFound)
	require.Equal(t, "INVALID_INPUT", apiv1.ErrCodeInvalidInput)
	require.Equal(t, "INTERNAL", apiv1.ErrCodeInternal)
}
