//go:build unit

package apiv1_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

func TestContractAliasesRoundTripJSON(t *testing.T) {
	t.Parallel()

	req := apiv1.CreateItemRequest{Name: "from-a-consumer"}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"from-a-consumer"}`, string(data))

	resp := apiv1.ItemResponse{ID: "id", Name: "n", CreatedAt: "t1", UpdatedAt: "t2"}
	data, err = json.Marshal(apiv1.ListItemsResponse{Items: []apiv1.ItemResponse{resp}})
	require.NoError(t, err)
	require.JSONEq(t, `{"items":[{"id":"id","name":"n","created_at":"t1","updated_at":"t2"}]}`, string(data))
}

func TestErrCodeReExports(t *testing.T) {
	t.Parallel()

	require.Equal(t, "NOT_FOUND", apiv1.ErrCodeNotFound)
	require.Equal(t, "INVALID_INPUT", apiv1.ErrCodeInvalidInput)
	require.Equal(t, "INTERNAL", apiv1.ErrCodeInternal)
}
