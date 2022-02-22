package nodeclient_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestOutputsQuery_Build(t *testing.T) {
	query := &nodeclient.OutputsQuery{
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
	hexTxID := hex.EncodeToString(txID[:])

	outputRes := &nodeclient.OutputResponse{
		TransactionID: hexTxID,
		OutputIndex:   3,
		Spent:         true,
		LedgerIndex:   1337,
		RawOutput:     &rawMsgSigDepJson,
	}

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

	client := nodeclient.New(nodeAPIUrl, iotago.ZeroRentParas, nodeclient.WithIndexer())

	resultSet, err := client.Indexer().Outputs(context.TODO(), &nodeclient.OutputsQuery{Tag: "some-tag"})
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
