package nodeclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/nodeclient"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestOutputsQuery_Build(t *testing.T) {
	trueCondition := true
	query := &apimodels.BasicOutputsQuery{
		IndexerTimelockParams: apimodels.IndexerTimelockParams{
			HasTimelock:      &trueCondition,
			TimelockedBefore: 1,
			TimelockedAfter:  2,
		},
		IndexerExpirationParams: apimodels.IndexerExpirationParams{
			HasExpiration: &trueCondition,
			ExpiresBefore: 5,
			ExpiresAfter:  6,
		},
		IndexerCreationParams: apimodels.IndexerCreationParams{
			CreatedBefore: 9,
			CreatedAfter:  10,
		},
		IndexerStorageDepositParams: apimodels.IndexerStorageDepositParams{
			HasStorageDepositReturn:           &trueCondition,
			StorageDepositReturnAddressBech32: "",
		},
		AddressBech32: "alice",
		SenderBech32:  "bob",
		Tag:           "charlie",
		IndexerCursorParams: apimodels.IndexerCursorParams{
			Cursor: func() *string {
				str := "dave"

				return &str
			}(),
		},
	}

	_, err := query.URLParams()
	require.NoError(t, err)
}

func Test_IndexerEnabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{"indexer/v2"},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Indexer(context.TODO())
	require.NoError(t, err)
}

func Test_IndexerDisabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{"someplugin/v1"},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Indexer(context.TODO())
	require.ErrorIs(t, err, nodeclient.ErrIndexerPluginNotAvailable)
}

func TestIndexerClient_BasicOutputs(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)
	originOutputProof, err := iotago.NewOutputIDProof(tpkg.TestAPI, tpkg.Rand32ByteArray(), tpkg.RandSlot(), iotago.TxEssenceOutputs{originOutput}, 0)
	require.NoError(t, err)

	fakeOutputID, err := originOutputProof.OutputID(originOutput)
	require.NoError(t, err)

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{"indexer/v2"},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	mockGetJSONWithParams(nodeclient.IndexerAPIRouteBasicOutputs, 200, &apimodels.IndexerResponse{
		CommittedSlot: 1337,
		PageSize:      1,
		Items:         iotago.HexOutputIDs{fakeOutputID.ToHex()},
		Cursor:        "some-offset-key",
	}, map[string]string{
		"tag": "some-tag",
	})

	mockGetJSONWithParams(nodeclient.IndexerAPIRouteBasicOutputs, 200, &apimodels.IndexerResponse{
		CommittedSlot: 1338,
		PageSize:      1,
		Items:         iotago.HexOutputIDs{fakeOutputID.ToHex()},
	}, map[string]string{
		"cursor": "some-offset-key",
		"tag":    "some-tag",
	})

	outputRoute := fmt.Sprintf(nodeclient.RouteOutput, fakeOutputID.ToHex())
	mockGetBinary(outputRoute, 200, &apimodels.OutputResponse{
		Output:        originOutput,
		OutputIDProof: originOutputProof,
	}, true)

	client := nodeClient(t)

	indexer, err := client.Indexer(context.TODO())
	require.NoError(t, err)

	resultSet, err := indexer.Outputs(context.TODO(), &apimodels.BasicOutputsQuery{Tag: "some-tag"})
	require.NoError(t, err)

	var runs int
	for resultSet.Next() {
		runs++
		outputs, err := resultSet.Outputs(context.TODO())
		require.NoError(t, err)

		require.Equal(t, originOutput, outputs[0])
	}

	require.NoError(t, resultSet.Error)
	require.Equal(t, 2, runs)
}
