package apimodels_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

func Test_RoutesResponse(t *testing.T) {
	api := testAPI()
	{
		response := &apimodels.RoutesResponse{
			Routes: []string{"route1", "route2"},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"routes\":[\"route1\",\"route2\"]}"
		require.Equal(t, expected, string(jsonResponse))
	}
}
