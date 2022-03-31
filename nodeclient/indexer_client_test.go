package nodeclient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestOutputsQuery_Build(t *testing.T) {
	query := &nodeclient.BasicOutputsQuery{
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelockCondition:      true,
			TimelockedBefore:          1,
			TimelockedAfter:           2,
			TimelockedBeforeMilestone: 3,
			TimelockedAfterMilestone:  4,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpirationCondition: true,
			ExpiresBefore:          5,
			ExpiresAfter:           6,
			ExpiresBeforeMilestone: 7,
			ExpiresAfterMilestone:  8,
		},
		IndexerCreationParas: nodeclient.IndexerCreationParas{
			CreatedBefore: 9,
			CreatedAfter:  10,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			RequiresStorageDepositReturn:      true,
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

	originInfo := &nodeclient.InfoResponse{
		Plugins: []string{"indexer/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.NodeAPIRouteInfo).
		Reply(200).
		JSON(originInfo)

	client := nodeclient.New(nodeAPIUrl)

	_, err := client.Indexer(context.TODO())
	require.NoError(t, err)
}

func Test_IndexerDisabled(t *testing.T) {
	defer gock.Off()

	originInfo := &nodeclient.InfoResponse{
		Plugins: []string{"someplugin/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.NodeAPIRouteInfo).
		Reply(200).
		JSON(originInfo)

	client := nodeclient.New(nodeAPIUrl)

	_, err := client.Indexer(context.TODO())
	require.ErrorIs(t, err, nodeclient.ErrIndexerPluginNotAvailable)
}

func TestIndexerClient_Outputs(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)
	originOutput.NativeTokens = iotago.NativeTokens{}
	originOutput.Blocks = iotago.FeatureBlocks{}
	sigDepJson, err := originOutput.MarshalJSON()
	require.NoError(t, err)
	rawMsgSigDepJson := json.RawMessage(sigDepJson)

	txID := tpkg.Rand32ByteArray()
	fakeOutputID := iotago.OutputIDFromTransactionIDAndIndex(txID, 1).ToHex()
	hexTxID := iotago.EncodeHex(txID[:])

	outputRes := &nodeclient.OutputResponse{
		TransactionID: hexTxID,
		OutputIndex:   3,
		Spent:         true,
		LedgerIndex:   1337,
		RawOutput:     &rawMsgSigDepJson,
	}

	originInfo := &nodeclient.InfoResponse{
		Plugins: []string{"indexer/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.NodeAPIRouteInfo).
		Reply(200).
		JSON(originInfo)

	gock.New(nodeAPIUrl).
		Get(nodeclient.IndexerAPIRouteOutputs).
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
		Get(nodeclient.IndexerAPIRouteOutputs).
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

	outputRoute := fmt.Sprintf(nodeclient.NodeAPIRouteOutput, fakeOutputID)
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
