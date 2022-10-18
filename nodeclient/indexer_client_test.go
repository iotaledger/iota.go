package nodeclient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestOutputsQuery_Build(t *testing.T) {
	trueCondition := true
	query := &nodeclient.BasicOutputsQuery{
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock:      &trueCondition,
			TimelockedBefore: 1,
			TimelockedAfter:  2,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &trueCondition,
			ExpiresBefore: 5,
			ExpiresAfter:  6,
		},
		IndexerCreationParas: nodeclient.IndexerCreationParas{
			CreatedBefore: 9,
			CreatedAfter:  10,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn:           &trueCondition,
			StorageDepositReturnAddressBech32: "",
		},
		AddressBech32: "alice",
		SenderBech32:  "bob",
		Tag:           "charlie",
		IndexerCursorParas: nodeclient.IndexerCursorParas{
			Cursor: func() *string {
				str := "dave"
				return &str
			}(),
		},
	}

	_, err := query.URLParas()
	require.NoError(t, err)
}

func Test_IndexerEnabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &nodeclient.RoutesResponse{
		Routes: []string{"indexer/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteRoutes).
		Reply(200).
		JSON(originRoutes)

	client := nodeclient.New(nodeAPIUrl)

	_, err := client.Indexer(context.TODO())
	require.NoError(t, err)
}

func Test_IndexerDisabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &nodeclient.RoutesResponse{
		Routes: []string{"someplugin/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteRoutes).
		Reply(200).
		JSON(originRoutes)

	client := nodeclient.New(nodeAPIUrl)

	_, err := client.Indexer(context.TODO())
	require.ErrorIs(t, err, nodeclient.ErrIndexerPluginNotAvailable)
}

func TestIndexerClient_BasicOutputs(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)
	sigDepJson, err := v2API.JSONEncode(originOutput)
	require.NoError(t, err)
	rawMsgSigDepJson := json.RawMessage(sigDepJson)

	txID := tpkg.Rand32ByteArray()
	fakeOutputID := iotago.OutputIDFromTransactionIDAndIndex(txID, 1).ToHex()
	hexTxID := iotago.EncodeHex(txID[:])

	outputRes := &nodeclient.OutputResponse{
		Metadata: &nodeclient.OutputMetadataResponse{
			TransactionID: hexTxID,
			OutputIndex:   3,
			Spent:         true,
			LedgerIndex:   1337,
		},
		RawOutput: &rawMsgSigDepJson,
	}

	originRoutes := &nodeclient.RoutesResponse{
		Routes: []string{"indexer/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteRoutes).
		Reply(200).
		JSON(originRoutes)

	gock.New(nodeAPIUrl).
		Get(nodeclient.IndexerAPIRouteBasicOutputs).
		MatchParam("tag", "some-tag").
		Reply(200).
		JSON(nodeclient.IndexerResponse{
			LedgerIndex: 1337,
			PageSize:    1,
			Items:       iotago.HexOutputIDs{fakeOutputID},
			Cursor: func() *string {
				str := "some-offset-key"
				return &str
			}(),
		})

	gock.New(nodeAPIUrl).
		Get(nodeclient.IndexerAPIRouteBasicOutputs).
		MatchParams(map[string]string{
			"cursor": "some-offset-key",
			"tag":    "some-tag",
		}).
		Reply(200).
		JSON(nodeclient.IndexerResponse{
			LedgerIndex: 1338,
			PageSize:    1,
			Items:       iotago.HexOutputIDs{fakeOutputID},
			Cursor:      nil,
		})

	outputRoute := fmt.Sprintf(nodeclient.RouteOutput, fakeOutputID)
	gock.New(nodeAPIUrl).
		Persist().
		Get(outputRoute).
		Reply(200).
		JSON(outputRes)

	client := nodeclient.New(nodeAPIUrl)

	indexer, err := client.Indexer(context.TODO())
	require.NoError(t, err)

	resultSet, err := indexer.Outputs(context.TODO(), &nodeclient.BasicOutputsQuery{Tag: "some-tag"})
	require.NoError(t, err)

	var runs int
	for resultSet.Next() {
		runs++
		outputs, err := resultSet.Outputs()
		require.NoError(t, err)

		require.Equal(t, originOutput, outputs[0])
	}

	require.NoError(t, resultSet.Error)
	require.Equal(t, 2, runs)
}
