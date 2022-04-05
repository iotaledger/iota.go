package nodeclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

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
			IsHealthy:                true,
			LatestMilestoneTimestamp: 1333337,
			LatestMilestoneIndex:     1337,
			ConfirmedMilestoneIndex:  666,
			PruningIndex:             142857,
		},
		Protocol: nodeclient.InfoResProtocol{
			NetworkName: "alphanet",
			Bech32HRP:   "atoi",
			MinPowScore: 4000.0,
			RentStructure: iotago.RentStructure{
				VByteCost:    500,
				VBFactorData: 1,
				VBFactorKey:  10,
			},
		},
		Metrics: nodeclient.InfoResMetrics{
			MessagesPerSecond:           20.0,
			ReferencedMessagesPerSecond: 10.0,
			ReferencedRate:              50.0,
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

func TestClient_SubmitMessage(t *testing.T) {
	defer gock.Off()

	msgHash := tpkg.Rand32ByteArray()
	msgHashStr := iotago.EncodeHex(msgHash[:])

	incompleteMsg := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(1),
	}

	completeMsg := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
		Payload: nil,
		Nonce:   3495721389537486,
	}

	serializedCompleteMsg, err := completeMsg.Serialize(serializer.DeSeriModeNoValidation, iotago.ZeroRentParas)
	require.NoError(t, err)

	// we need to do this, otherwise gock doesn't match the body
	gock.BodyTypes = append(gock.BodyTypes, "application/octet-stream")
	gock.BodyTypeAliases["octet"] = "application/octet-stream"

	serializedIncompleteMsg, err := incompleteMsg.Serialize(serializer.DeSeriModePerformValidation, iotago.ZeroRentParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Post(nodeclient.RouteMessages).
		MatchType("octet").
		Body(bytes.NewReader(serializedIncompleteMsg)).
		Reply(200).
		AddHeader("Location", msgHashStr)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMessageBytes, msgHashStr)).
		Reply(200).
		Body(bytes.NewReader(serializedCompleteMsg))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.SubmitMessage(context.Background(), incompleteMsg, iotago.ZeroRentParas)
	require.NoError(t, err)
	require.EqualValues(t, completeMsg, resp)
}

func TestClient_MessageMetadataByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	parents := tpkg.SortedRand32BytArray(1 + rand.Intn(7))

	queryHash := iotago.EncodeHex(identifier[:])

	parentMessageIDs := make([]string, len(parents))
	for i, p := range parents {
		parentMessageIDs[i] = iotago.EncodeHex(p[:])
	}

	originRes := &nodeclient.MessageMetadataResponse{
		MessageID:                  queryHash,
		Parents:                    parentMessageIDs,
		Solid:                      true,
		MilestoneIndex:             nil,
		ReferencedByMilestoneIndex: nil,
		LedgerInclusionState:       nil,
		ShouldPromote:              nil,
		ShouldReattach:             nil,
		ConflictReason:             0,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMessageMetadata, queryHash)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	meta, err := nodeAPI.MessageMetadataByMessageID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originRes, meta)
}

func TestClient_MessageByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	queryHash := iotago.EncodeHex(identifier[:])

	originMsg := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
		Payload: nil,
		Nonce:   16345984576234,
	}

	data, err := originMsg.Serialize(serializer.DeSeriModePerformValidation, iotago.ZeroRentParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMessageBytes, queryHash)).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	responseMsg, err := nodeAPI.MessageByMessageID(context.Background(), identifier, iotago.ZeroRentParas)
	require.NoError(t, err)
	require.EqualValues(t, originMsg, responseMsg)
}

func TestClient_ChildrenByMessageID(t *testing.T) {
	defer gock.Off()

	msgID := tpkg.Rand32ByteArray()
	hexMsgID := iotago.EncodeHex(msgID[:])

	child1 := tpkg.Rand32ByteArray()
	child2 := tpkg.Rand32ByteArray()
	child3 := tpkg.Rand32ByteArray()

	originRes := &nodeclient.ChildrenResponse{
		MessageID:  hexMsgID,
		MaxResults: 1000,
		Count:      3,
		Children: []string{
			iotago.EncodeHex(child1[:]),
			iotago.EncodeHex(child2[:]),
			iotago.EncodeHex(child3[:]),
		},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMessageChildren, hexMsgID)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	res, err := nodeAPI.ChildrenByMessageID(context.Background(), msgID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_TransactionIncludedMessage(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	queryHash := iotago.EncodeHex(identifier[:])

	originMsg := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
		Payload: nil,
		Nonce:   16345984576234,
	}

	data, err := originMsg.Serialize(serializer.DeSeriModePerformValidation, iotago.ZeroRentParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteTransactionsIncludedMessage, queryHash)).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeclient.New(nodeAPIUrl)
	responseMsg, err := nodeAPI.TransactionIncludedMessage(context.Background(), identifier, iotago.ZeroRentParas)
	require.NoError(t, err)
	require.EqualValues(t, originMsg, responseMsg)
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
		TransactionID: hexTxID,
		OutputIndex:   3,
		Spent:         true,
		LedgerIndex:   1337,
		RawOutput:     &rawMsgSigDepJson,
	}

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteOutput, utxoInputId.ToHex())).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.OutputByID(context.Background(), utxoInputId)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	resTxID, err := resp.TxID()
	require.NoError(t, err)
	require.EqualValues(t, txID, *resTxID)
}

func TestClient_OutputMetadataByID(t *testing.T) {
	defer gock.Off()

	txID := tpkg.Rand32ByteArray()
	hexTxID := iotago.EncodeHex(txID[:])
	originRes := &nodeclient.OutputResponse{
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
	resp, err := nodeAPI.OutputByID(context.Background(), utxoInputId)
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
				Receipt: &iotago.Receipt{
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
				Receipt: &iotago.Receipt{
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

func TestClient_MilestoneByIndex(t *testing.T) {
	defer gock.Off()

	var milestoneIndex uint32 = 1337
	msgID := tpkg.Rand32ByteArray()

	originRes := &nodeclient.MilestoneResponse{
		Index:     milestoneIndex,
		MessageID: iotago.EncodeHex(msgID[:]),
		Time:      time.Now().Unix(),
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteMilestone, milestoneIndex)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.MilestoneByIndex(context.Background(), milestoneIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
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
		Get(fmt.Sprintf(nodeclient.RouteMilestoneUTXOChanges, milestoneIndex)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeclient.New(nodeAPIUrl)
	resp, err := nodeAPI.MilestoneUTXOChangesByIndex(context.Background(), milestoneIndex)
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
		NewMessages:               40,
		KnownMessages:             60,
		ReceivedMessages:          100,
		ReceivedMessageRequests:   345,
		ReceivedMilestoneRequests: 194,
		ReceivedHeartbeats:        5,
		SentMessages:              492,
		SentMessageRequests:       2396,
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
