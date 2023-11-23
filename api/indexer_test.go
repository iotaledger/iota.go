package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
)

func Test_IndexerResponse(t *testing.T) {
	testAPI := testAPI()
	{
		response := &api.IndexerResponse{
			CommittedSlot: 281,
			PageSize:      1000,
			Items: iotago.OutputIDs{
				iotago.OutputID{0xff},
				iotago.OutputID{0xfa},
			}.ToHex(),
			Cursor: "cursor-value",
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"committedSlot\":281,\"pageSize\":1000,\"items\":[\"0xff00000000000000000000000000000000000000000000000000000000000000000000000000\",\"0xfa00000000000000000000000000000000000000000000000000000000000000000000000000\"],\"cursor\":\"cursor-value\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.IndexerResponse)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &api.IndexerResponse{
			CommittedSlot: 281,
			PageSize:      1000,
			Items: iotago.OutputIDs{
				iotago.OutputID{0xff},
				iotago.OutputID{0xfa},
			}.ToHex(),
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"committedSlot\":281,\"pageSize\":1000,\"items\":[\"0xff00000000000000000000000000000000000000000000000000000000000000000000000000\",\"0xfa00000000000000000000000000000000000000000000000000000000000000000000000000\"]}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.IndexerResponse)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}
