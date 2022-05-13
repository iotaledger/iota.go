package nodeclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3/nodeclient"
)

const nodeAPIUrl = "http://127.0.0.1:14265"

func TestClient_Health(t *testing.T) {
	defer gock.Off()
	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteHealth).
		Reply(200)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	healthy, err := nodeAPI.Health(context.Background())
	require.NoError(t, err)
	require.True(t, healthy)

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteHealth).
		Reply(503)

	healthy, err = nodeAPI.Health(context.Background())
	require.NoError(t, err)
	require.False(t, healthy)
}

func TestClient_Info(t *testing.T) {
	defer gock.Off()

	originInfo := &nodeclient.InfoResponse{
		Name:    "HORNET",
		Version: "1.0.0",
		Status: nodeclient.InfoResStatus{
			IsHealthy: true,
			LatestMilestone: nodeclient.InfoResMilestone{
				Index:       1337,
				Timestamp:   1333337,
				MilestoneID: iotago.EncodeHex(tpkg.RandBytes(32)),
			},
			ConfirmedMilestone: nodeclient.InfoResMilestone{
				Index:       666,
				Timestamp:   6666,
				MilestoneID: iotago.EncodeHex(tpkg.RandBytes(32)),
			},
			PruningIndex: 142857,
		},
		Protocol: iotago.ProtocolParameters{
			TokenSupply: tpkg.TestTokenSupply,
			Version:     2,
			NetworkName: "alphanet",
			Bech32HRP:   "atoi",
			MinPoWScore: 40000.0,
			RentStructure: iotago.RentStructure{
				VByteCost:    500,
				VBFactorData: 1,
				VBFactorKey:  10,
			},
		},
		BaseToken: &nodeclient.InfoResBaseToken{
			Name:            "TestCoin",
			TickerSymbol:    "TEST",
			Unit:            "TEST",
			Subunit:         "testies",
			Decimals:        6,
			UseMetricPrefix: false,
		},
		Metrics: nodeclient.InfoResMetrics{
			BlocksPerSecond:           20.0,
			ReferencedBlocksPerSecond: 10.0,
			ReferencedRate:            50.0,
		},
		Features: []string{"Lazers"},
		Plugins:  []string{"indexer/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteInfo).
		Reply(200).
		JSON(originInfo)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	info, err := nodeAPI.Info(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originInfo, info)
	protoParas := originInfo.Protocol
	require.NoError(t, err)
	require.EqualValues(t, protoParas.TokenSupply, tpkg.TestTokenSupply)
}

func TestClient_Tips(t *testing.T) {
	defer gock.Off()

	originRes := &nodeclient.TipsResponse{
		TipsHex: []string{"733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d", "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteTips).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	tips, err := nodeAPI.Tips(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, tips)
}

func TestClient_SubmitBlock(t *testing.T) {
	defer gock.Off()

	blockHash := tpkg.Rand32ByteArray()
	blockHashStr := iotago.EncodeHex(blockHash[:])

	incompleteBlock := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion,
		Parents:         tpkg.SortedRand32BytArray(1),
	}

	completeBlock := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion,
		Parents:         tpkg.SortedRand32BytArray(1),
		Payload:         nil,
		Nonce:           3495721389537486,
	}

	serializedCompleteBlock, err := completeBlock.Serialize(serializer.DeSeriModeNoValidation, tpkg.TestProtoParas)
	require.NoError(t, err)

	block2 := iotago.Block{}
	_, err = block2.Deserialize(serializedCompleteBlock, serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	require.NoError(t, err)

	serializedIncompleteBlock, err := incompleteBlock.Serialize(serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Post(nodeclient.RouteBlocks).
		MatchType(nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Body(bytes.NewReader(serializedIncompleteBlock)).
		Reply(200).
		AddHeader("Location", blockHashStr)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteBlock, blockHashStr)).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(serializedCompleteBlock))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.SubmitBlock(context.Background(), incompleteBlock, tpkg.TestProtoParas)
	require.NoError(t, err)
	require.EqualValues(t, completeBlock, resp)
}

func TestClient_BlockMetadataByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	parents := tpkg.SortedRand32BytArray(1 + rand.Intn(7))

	queryHash := iotago.EncodeHex(identifier[:])

	parentBlockIDs := make([]string, len(parents))
	for i, p := range parents {
		parentBlockIDs[i] = iotago.EncodeHex(p[:])
	}

	originRes := &nodeclient.BlockMetadataResponse{
		BlockID:                    queryHash,
		Parents:                    parentBlockIDs,
		Solid:                      true,
		MilestoneIndex:             nil,
		ReferencedByMilestoneIndex: nil,
		LedgerInclusionState:       nil,
		ShouldPromote:              nil,
		ShouldReattach:             nil,
		ConflictReason:             0,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteBlockMetadata, queryHash)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	meta, err := nodeAPI.BlockMetadataByBlockID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originRes, meta)
}

func TestClient_BlockByBlockID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	queryHash := iotago.EncodeHex(identifier[:])

	originBlock := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion,
		Parents:         tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
		Payload:         nil,
		Nonce:           16345984576234,
	}

	data, err := originBlock.Serialize(serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteBlock, queryHash)).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	responseBlock, err := nodeAPI.BlockByBlockID(context.Background(), identifier, tpkg.TestProtoParas)
	require.NoError(t, err)
	require.EqualValues(t, originBlock, responseBlock)
}

func TestClient_ChildrenByBlockID(t *testing.T) {
	defer gock.Off()

	blockID := tpkg.Rand32ByteArray()
	hexBlockID := iotago.EncodeHex(blockID[:])

	child1 := tpkg.Rand32ByteArray()
	child2 := tpkg.Rand32ByteArray()
	child3 := tpkg.Rand32ByteArray()

	originRes := &nodeclient.ChildrenResponse{
		BlockID:    hexBlockID,
		MaxResults: 1000,
		Count:      3,
		Children: []string{
			iotago.EncodeHex(child1[:]),
			iotago.EncodeHex(child2[:]),
			iotago.EncodeHex(child3[:]),
		},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteBlockChildren, hexBlockID)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	res, err := nodeAPI.ChildrenByBlockID(context.Background(), blockID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_TransactionIncludedBlock(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	queryHash := iotago.EncodeHex(identifier[:])

	originBlock := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion,
		Parents:         tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
		Payload:         nil,
		Nonce:           16345984576234,
	}

	data, err := originBlock.Serialize(serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteTransactionsIncludedBlock, queryHash)).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	responseBlock, err := nodeAPI.TransactionIncludedBlock(context.Background(), identifier, tpkg.TestProtoParas)
	require.NoError(t, err)
	require.EqualValues(t, originBlock, responseBlock)
}

func TestClient_OutputByID(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)
	sigDepJson, err := originOutput.MarshalJSON()
	require.NoError(t, err)
	rawMsgSigDepJson := json.RawMessage(sigDepJson)

	txID := tpkg.Rand32ByteArray()
	hexTxID := iotago.EncodeHex(txID[:])
	originRes := &nodeclient.OutputResponse{
		Metadata: &nodeclient.OutputMetadataResponse{
			TransactionID: hexTxID,
			OutputIndex:   3,
			Spent:         true,
			LedgerIndex:   1337,
		},
		RawOutput: &rawMsgSigDepJson,
	}

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteOutput, utxoInputId.ToHex())).
		MatchHeader("Accept", nodeclient.MIMEApplicationJSON).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.OutputByID(context.Background(), utxoInputId)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	resTxID, err := resp.Metadata.TxID()
	require.NoError(t, err)
	require.EqualValues(t, txID, *resTxID)
}

func TestClient_OutputMetadataByID(t *testing.T) {
	defer gock.Off()

	txID := tpkg.Rand32ByteArray()
	hexTxID := iotago.EncodeHex(txID[:])
	originRes := &nodeclient.OutputMetadataResponse{
		TransactionID: hexTxID,
		OutputIndex:   3,
		Spent:         true,
		LedgerIndex:   1337,
	}

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteOutput, utxoInputId.ToHex())).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.OutputMetadataByID(context.Background(), utxoInputId)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	resTxID, err := resp.TxID()
	require.NoError(t, err)
	require.EqualValues(t, txID, *resTxID)
}

func TestNodeHTTPAPIClient_Treasury(t *testing.T) {
	defer gock.Off()

	originRes := &nodeclient.TreasuryResponse{
		MilestoneID: "0x733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		Amount:      "133713371337",
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteTreasury).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.Treasury(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeHTTPAPIClient_Receipts(t *testing.T) {
	defer gock.Off()

	originRes := &nodeclient.ReceiptsResponse{
		Receipts: []*nodeclient.ReceiptTuple{
			{
				MilestoneIndex: 1000,
				Receipt: &iotago.ReceiptMilestoneOpt{
					MigratedAt: 1000,
					Final:      false,
					Funds: iotago.MigratedFundsEntries{
						&iotago.MigratedFundsEntry{
							TailTransactionHash: iotago.LegacyTailTransactionHash{},
							Address:             &iotago.Ed25519Address{},
							Deposit:             10000,
						},
					},
					Transaction: &iotago.TreasuryTransaction{
						Input:  &iotago.TreasuryInput{},
						Output: &iotago.TreasuryOutput{Amount: 10000},
					},
				},
			},
		},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteReceipts).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.Receipts(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes.Receipts, resp)
}

func TestNodeHTTPAPIClient_ReceiptsByMigratedAtIndex(t *testing.T) {
	defer gock.Off()

	var index uint32 = 1000

	originRes := &nodeclient.ReceiptsResponse{
		Receipts: []*nodeclient.ReceiptTuple{
			{
				MilestoneIndex: 1000,
				Receipt: &iotago.ReceiptMilestoneOpt{
					MigratedAt: 1000,
					Final:      false,
					Funds: iotago.MigratedFundsEntries{
						&iotago.MigratedFundsEntry{
							TailTransactionHash: iotago.LegacyTailTransactionHash{},
							Address:             &iotago.Ed25519Address{},
							Deposit:             10000,
						},
					},
					Transaction: &iotago.TreasuryTransaction{
						Input:  &iotago.TreasuryInput{},
						Output: &iotago.TreasuryOutput{Amount: 10000},
					},
				},
			},
		},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteReceiptsMigratedAtIndex, index)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.ReceiptsByMigratedAtIndex(context.Background(), index)
	require.NoError(t, err)
	require.EqualValues(t, originRes.Receipts, resp)
}

func TestClient_MilestoneByID(t *testing.T) {
	defer gock.Off()

	milestoneID := tpkg.RandMilestoneID()

	milestone := &iotago.Milestone{
		Index:               1337,
		Timestamp:           1337,
		PreviousMilestoneID: tpkg.RandMilestoneID(),
		Parents: iotago.MilestoneParentBlockIDs{
			tpkg.Rand32ByteArray(),
		},
		ConfirmedMerkleRoot: tpkg.Rand32ByteArray(),
		AppliedMerkleRoot:   tpkg.Rand32ByteArray(),
		Metadata:            tpkg.RandBytes(30),
		Opts: iotago.MilestoneOpts{
			&iotago.ProtocolParamsMilestoneOpt{
				TargetMilestoneIndex: 500,
				ProtocolVersion:      2,
				Params:               []byte{1, 2, 3, 4, 5, 6, 7},
			},
		},
		Signatures: iotago.Signatures{
			tpkg.RandEd25519Signature(),
		},
	}

	data, err := milestone.Serialize(serializer.DeSeriModeNoValidation, tpkg.TestProtoParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMilestoneByID, iotago.EncodeHex(milestoneID[:]))).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.MilestoneByID(context.Background(), milestoneID)
	require.NoError(t, err)
	require.EqualValues(t, milestone, resp)
}

func TestClient_MilestoneUTXOChangesByID(t *testing.T) {
	defer gock.Off()

	milestoneID := tpkg.RandMilestoneID()

	randCreatedOutput := tpkg.RandUTXOInput()
	randConsumedOutput := tpkg.RandUTXOInput()

	originRes := &nodeclient.MilestoneUTXOChangesResponse{
		Index:           1337,
		CreatedOutputs:  []string{randCreatedOutput.ID().ToHex()},
		ConsumedOutputs: []string{randConsumedOutput.ID().ToHex()},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMilestoneByIDUTXOChanges, iotago.EncodeHex(milestoneID[:]))).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.MilestoneUTXOChangesByID(context.Background(), milestoneID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_MilestoneByIndex(t *testing.T) {
	defer gock.Off()

	var milestoneIndex uint32 = 1337

	milestone := &iotago.Milestone{
		Index:               milestoneIndex,
		Timestamp:           1337,
		PreviousMilestoneID: tpkg.RandMilestoneID(),
		Parents: iotago.MilestoneParentBlockIDs{
			tpkg.Rand32ByteArray(),
		},
		ConfirmedMerkleRoot: tpkg.Rand32ByteArray(),
		AppliedMerkleRoot:   tpkg.Rand32ByteArray(),
		Metadata:            tpkg.RandBytes(30),
		Opts: iotago.MilestoneOpts{
			&iotago.ProtocolParamsMilestoneOpt{
				TargetMilestoneIndex: 500,
				ProtocolVersion:      2,
				Params:               []byte{1, 2, 3, 4, 5, 6, 7},
			},
		},
		Signatures: iotago.Signatures{
			tpkg.RandEd25519Signature(),
		},
	}

	data, err := milestone.Serialize(serializer.DeSeriModeNoValidation, nil)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMilestoneByIndex, milestoneIndex)).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.MilestoneByIndex(context.Background(), milestoneIndex)
	require.NoError(t, err)
	require.EqualValues(t, milestone, resp)
}

func TestClient_MilestoneUTXOChangesByIndex(t *testing.T) {
	defer gock.Off()

	var milestoneIndex uint32 = 1337

	randCreatedOutput := tpkg.RandUTXOInput()
	randConsumedOutput := tpkg.RandUTXOInput()

	originRes := &nodeclient.MilestoneUTXOChangesResponse{
		Index:           milestoneIndex,
		CreatedOutputs:  []string{randCreatedOutput.ID().ToHex()},
		ConsumedOutputs: []string{randConsumedOutput.ID().ToHex()},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMilestoneByIndexUTXOChanges, milestoneIndex)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.MilestoneUTXOChangesByIndex(context.Background(), milestoneIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_ComputeWhiteFlagMutations(t *testing.T) {
	defer gock.Off()

	var milestoneIndex uint32 = 1337
	var milestoneTimestamp uint32 = 1333337

	parents := tpkg.SortedRand32BytArray(1 + rand.Intn(7))
	parentBlockIDs := make([]string, len(parents))
	for i, p := range parents {
		parentBlockIDs[i] = iotago.EncodeHex(p[:])
	}

	randConfirmedMerkleRoot := tpkg.RandMilestoneMerkleProof()
	randAppliedMerkleRoot := tpkg.RandMilestoneMerkleProof()

	milestoneID := tpkg.RandMilestoneID()
	req := &nodeclient.ComputeWhiteFlagMutationsRequest{
		Index:               milestoneIndex,
		Timestamp:           milestoneTimestamp,
		Parents:             parentBlockIDs,
		PreviousMilestoneID: iotago.EncodeHex(milestoneID[:]),
	}

	internalRes := &nodeclient.ComputeWhiteFlagMutationsResponseInternal{
		ConfirmedMerkleRoot: iotago.EncodeHex(randConfirmedMerkleRoot[:]),
		AppliedMerkleRoot:   iotago.EncodeHex(randAppliedMerkleRoot[:]),
	}

	originRes := &nodeclient.ComputeWhiteFlagMutationsResponse{
		ConfirmedMerkleRoot: randConfirmedMerkleRoot,
		AppliedMerkleRoot:   randAppliedMerkleRoot,
	}

	gock.New(nodeAPIUrl).
		Post(nodeclient.RouteComputeWhiteFlagMutations).
		JSON(req).
		Reply(200).
		JSON(internalRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.ComputeWhiteFlagMutations(context.Background(), milestoneIndex, milestoneTimestamp, parents, milestoneID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

var sampleGossipInfo = &nodeclient.GossipInfo{
	Heartbeat: &nodeclient.GossipHeartbeat{
		SolidMilestoneIndex:  234,
		PrunedMilestoneIndex: 5872,
		LatestMilestoneIndex: 1294,
		ConnectedNeighbors:   2392,
		SyncedNeighbors:      1234,
	},
	Metrics: nodeclient.PeerGossipMetrics{
		NewBlocks:                 40,
		KnownBlocks:               60,
		ReceivedBlocks:            100,
		ReceivedBlockRequests:     345,
		ReceivedMilestoneRequests: 194,
		ReceivedHeartbeats:        5,
		SentBlocks:                492,
		SentBlockRequests:         2396,
		SentMilestoneRequests:     9837,
		SentHeartbeats:            3,
		DroppedPackets:            10,
	},
}

func TestClient_PeerByID(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"

	originRes := &nodeclient.PeerResponse{
		MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		Gossip:         sampleGossipInfo,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RoutePeer, peerID)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.PeerByID(context.Background(), peerID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_RemovePeerByID(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"

	gock.New(nodeAPIUrl).
		Delete(fmt.Sprintf(nodeclient.RoutePeer, peerID)).
		Reply(200).
		Status(200)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	err := nodeAPI.RemovePeerByID(context.Background(), peerID)
	require.NoError(t, err)
}

func TestClient_Peers(t *testing.T) {
	defer gock.Off()

	peerID1 := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
	peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

	originRes := []*nodeclient.PeerResponse{
		{
			MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID1)},
			ID:             peerID1,
			Connected:      true,
			Relation:       "autopeered",
			Gossip:         sampleGossipInfo,
		},
		{
			MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID2)},
			ID:             peerID2,
			Connected:      true,
			Relation:       "static",
			Gossip:         sampleGossipInfo,
		},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RoutePeers).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.Peers(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_AddPeer(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
	multiAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)

	originRes := &nodeclient.PeerResponse{
		MultiAddresses: []string{multiAddr},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		Gossip:         sampleGossipInfo,
	}

	req := &nodeclient.AddPeerRequest{MultiAddress: multiAddr}
	gock.New(nodeAPIUrl).
		Post(nodeclient.RoutePeers).
		JSON(req).
		Reply(201).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.AddPeer(context.Background(), multiAddr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}
