package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v4/api"
)

func Test_RoutesResponse(t *testing.T) {
	testAPI := testAPI()
	{
		response := &api.RoutesResponse{
			Routes: []string{"route1", "route2"},
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"routes\":[\"route1\",\"route2\"]}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.RoutesResponse)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}
