package apimodels_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

func Test_IndexerResponse(t *testing.T) {
	api := testAPI()
	{
		response := &apimodels.IndexerResponse{
			LedgerIndex: 281,
			PageSize:    1000,
			Items: iotago.OutputIDs{
				iotago.OutputID{0xff},
				iotago.OutputID{0xfa},
			}.ToHex(),
			Cursor: "cursor-value",
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"ledgerIndex\":281,\"pageSize\":1000,\"items\":[\"0xff00000000000000000000000000000000000000000000000000000000000000000000000000\",\"0xfa00000000000000000000000000000000000000000000000000000000000000000000000000\"],\"cursor\":\"cursor-value\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.IndexerResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &apimodels.IndexerResponse{
			LedgerIndex: 281,
			PageSize:    1000,
			Items: iotago.OutputIDs{
				iotago.OutputID{0xff},
				iotago.OutputID{0xfa},
			}.ToHex(),
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"ledgerIndex\":281,\"pageSize\":1000,\"items\":[\"0xff00000000000000000000000000000000000000000000000000000000000000000000000000\",\"0xfa00000000000000000000000000000000000000000000000000000000000000000000000000\"]}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.IndexerResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}
