package iotago_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/v2/serializer"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/iotaledger/iota.go/v3/tpkg"

	iotago "github.com/iotaledger/iota.go/v3"
)

const nodeAPIUrl = "http://127.0.0.1:14265"

func TestNodeAPI_Health(t *testing.T) {
	defer gock.Off()
	gock.New(nodeAPIUrl).
		Get(iotago.NodeAPIRouteHealth).
		Reply(200)

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	healthy, err := nodeAPI.Health(context.Background())
	require.NoError(t, err)
	require.True(t, healthy)

	gock.New(nodeAPIUrl).
		Get(iotago.NodeAPIRouteHealth).
		Reply(503)

	healthy, err = nodeAPI.Health(context.Background())
	require.NoError(t, err)
	require.False(t, healthy)
}

func TestNodeAPI_Info(t *testing.T) {
	defer gock.Off()

	originInfo := &iotago.NodeInfoResponse{
		Name:                        "HORNET",
		Version:                     "1.0.0",
		IsHealthy:                   true,
		NetworkID:                   "alphanet@1",
		Bech32HRP:                   "atoi",
		MinPowScore:                 4000.0,
		MessagesPerSecond:           20.0,
		ReferencedMessagesPerSecond: 10.0,
		ReferencedRate:              50.0,
		LatestMilestoneTimestamp:    1333337,
		LatestMilestoneIndex:        1337,
		ConfirmedMilestoneIndex:     666,
		PruningIndex:                142857,
		Features:                    []string{"Lazers"},
	}

	gock.New(nodeAPIUrl).
		Get(iotago.NodeAPIRouteInfo).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originInfo})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	info, err := nodeAPI.Info(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originInfo, info)
}

func TestNodeAPI_Tips(t *testing.T) {
	defer gock.Off()

	originRes := &iotago.NodeTipsResponse{
		TipsHex: []string{"733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d", "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe"},
	}

	gock.New(nodeAPIUrl).
		Get(iotago.NodeAPIRouteTips).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	tips, err := nodeAPI.Tips(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, tips)
}

func TestNodeAPI_SubmitMessage(t *testing.T) {
	defer gock.Off()

	msgHash := tpkg.Rand32ByteArray()
	msgHashStr := hex.EncodeToString(msgHash[:])

	incompleteMsg := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(1),
	}

	completeMsg := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
		Payload: nil,
		Nonce:   3495721389537486,
	}

	serializedCompleteMsg, err := completeMsg.Serialize(serializer.DeSeriModeNoValidation, DefZeroRentParas)
	require.NoError(t, err)

	// we need to do this, otherwise gock doesn't match the body
	gock.BodyTypes = append(gock.BodyTypes, "application/octet-stream")
	gock.BodyTypeAliases["octet"] = "application/octet-stream"

	serializedIncompleteMsg, err := incompleteMsg.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Post(iotago.NodeAPIRouteMessages).
		MatchType("octet").
		Body(bytes.NewReader(serializedIncompleteMsg)).
		Reply(200).
		AddHeader("Location", msgHashStr)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRouteMessageBytes, msgHashStr)).
		Reply(200).
		Body(bytes.NewReader(serializedCompleteMsg))

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.SubmitMessage(context.Background(), incompleteMsg)
	require.NoError(t, err)
	require.EqualValues(t, completeMsg, resp)
}

func TestNodeAPI_MessageIDsByIndex(t *testing.T) {
	defer gock.Off()
	index := "बेकार पाठ"

	id1 := tpkg.Rand32ByteArray()
	id2 := tpkg.Rand32ByteArray()
	id3 := tpkg.Rand32ByteArray()

	msgIDsByIndex := &iotago.MessageIDsByIndexResponse{
		Index:      hex.EncodeToString([]byte(index)),
		MaxResults: 1000,
		Count:      3,
		MessageIDs: []string{
			hex.EncodeToString(id1[:]),
			hex.EncodeToString(id2[:]),
			hex.EncodeToString(id3[:]),
		},
	}

	gock.New(nodeAPIUrl).
		Get(iotago.NodeAPIRouteMessages).
		MatchParam("index", hex.EncodeToString([]byte(index))).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: msgIDsByIndex})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resMsgIDsByIndex, err := nodeAPI.MessageIDsByIndex(context.Background(), []byte(index))
	require.NoError(t, err)
	require.EqualValues(t, msgIDsByIndex, resMsgIDsByIndex)
}

func TestNodeAPI_MessageMetadataByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	parents := tpkg.SortedRand32BytArray(1 + rand.Intn(7))

	queryHash := hex.EncodeToString(identifier[:])

	parentMessageIDs := make([]string, len(parents))
	for i, p := range parents {
		parentMessageIDs[i] = hex.EncodeToString(p[:])
	}

	originRes := &iotago.MessageMetadataResponse{
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
		Get(fmt.Sprintf(iotago.NodeAPIRouteMessageMetadata, queryHash)).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	meta, err := nodeAPI.MessageMetadataByMessageID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originRes, meta)
}

func TestNodeAPI_MessageByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	queryHash := hex.EncodeToString(identifier[:])

	originMsg := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
		Payload: nil,
		Nonce:   16345984576234,
	}

	data, err := originMsg.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRouteMessageBytes, queryHash)).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	responseMsg, err := nodeAPI.MessageByMessageID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originMsg, responseMsg)
}

func TestNodeAPI_ChildrenByMessageID(t *testing.T) {
	defer gock.Off()

	msgID := tpkg.Rand32ByteArray()
	hexMsgID := hex.EncodeToString(msgID[:])

	child1 := tpkg.Rand32ByteArray()
	child2 := tpkg.Rand32ByteArray()
	child3 := tpkg.Rand32ByteArray()

	originRes := &iotago.ChildrenResponse{
		MessageID:  hexMsgID,
		MaxResults: 1000,
		Count:      3,
		Children: []string{
			hex.EncodeToString(child1[:]),
			hex.EncodeToString(child2[:]),
			hex.EncodeToString(child3[:]),
		},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRouteMessageChildren, hexMsgID)).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	res, err := nodeAPI.ChildrenByMessageID(context.Background(), msgID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestNodeAPI_OutputByID(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandExtendedOutput(iotago.AddressEd25519)
	sigDepJson, err := originOutput.MarshalJSON()
	require.NoError(t, err)
	rawMsgSigDepJson := json.RawMessage(sigDepJson)

	txID := tpkg.Rand32ByteArray()
	hexTxID := hex.EncodeToString(txID[:])
	originRes := &iotago.NodeOutputResponse{
		TransactionID: hexTxID,
		OutputIndex:   3,
		Spent:         true,
		LedgerIndex:   1337,
		RawOutput:     &rawMsgSigDepJson,
	}

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRouteOutput, utxoInputId.ToHex())).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.OutputByID(context.Background(), utxoInputId)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	resTxID, err := resp.TxID()
	require.NoError(t, err)
	require.EqualValues(t, txID, *resTxID)
}

func TestNodeAPI_BalanceByEd25519Address(t *testing.T) {
	defer gock.Off()

	ed25519Addr, _ := tpkg.RandEd25519AddressAndBytes()
	ed25519AddrHex := ed25519Addr.String()

	originRes := &iotago.AddressBalanceResponse{
		AddressType: 1,
		Address:     ed25519AddrHex,
		Balance:     13371337,
		LedgerIndex: 1337,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRouteAddressEd25519Balance, ed25519AddrHex)).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.BalanceByEd25519Address(context.Background(), ed25519Addr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_OutputIDsByAddress(t *testing.T) {
	defer gock.Off()

	ed25519Addr, _ := tpkg.RandEd25519AddressAndBytes()
	ed25519AddrHex := ed25519Addr.String()

	output1 := tpkg.Rand32ByteArray()
	output2 := tpkg.Rand32ByteArray()
	output3 := tpkg.Rand32ByteArray()
	originRes := &iotago.AddressOutputsResponse{
		AddressType: 1,
		Address:     ed25519AddrHex,
		MaxResults:  1000,
		Count:       2,
		OutputIDs: []iotago.OutputIDHex{
			iotago.OutputIDHex(hex.EncodeToString(output1[:])),
			iotago.OutputIDHex(hex.EncodeToString(output2[:])),
		},
		LedgerIndex: 1337,
	}

	originResWithUnspent := &iotago.AddressOutputsResponse{
		AddressType: 1,
		Address:     ed25519AddrHex,
		MaxResults:  1000,
		Count:       3,
		OutputIDs: []iotago.OutputIDHex{
			iotago.OutputIDHex(hex.EncodeToString(output1[:])),
			iotago.OutputIDHex(hex.EncodeToString(output2[:])),
			iotago.OutputIDHex(hex.EncodeToString(output3[:])),
		},
	}

	route := fmt.Sprintf(iotago.NodeAPIRouteAddressEd25519Outputs, ed25519AddrHex)
	gock.New(nodeAPIUrl).
		Get(route).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.OutputIDsByEd25519Address(context.Background(), ed25519Addr, false)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	gock.New(nodeAPIUrl).
		Get(route).
		MatchParam("include-spent", "true").
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originResWithUnspent})

	resp, err = nodeAPI.OutputIDsByEd25519Address(context.Background(), ed25519Addr, true)
	require.NoError(t, err)
	require.EqualValues(t, originResWithUnspent, resp)
}

func TestNodeHTTPAPIClient_Treasury(t *testing.T) {
	defer gock.Off()

	originRes := &iotago.TreasuryResponse{
		MilestoneID: "733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		Amount:      133713371337,
	}

	gock.New(nodeAPIUrl).
		Get(iotago.NodeAPIRouteTreasury).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.Treasury(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeHTTPAPIClient_Receipts(t *testing.T) {
	defer gock.Off()

	originRes := &iotago.ReceiptsResponse{
		Receipts: []*iotago.ReceiptTuple{
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
		Get(iotago.NodeAPIRouteReceipts).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.Receipts(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes.Receipts, resp)
}

func TestNodeHTTPAPIClient_ReceiptsByMigratedAtIndex(t *testing.T) {
	defer gock.Off()

	var index uint32 = 1000

	originRes := &iotago.ReceiptsResponse{
		Receipts: []*iotago.ReceiptTuple{
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
		Get(fmt.Sprintf(iotago.NodeAPIRouteReceiptsByMigratedAtIndex, strconv.FormatUint(uint64(index), 10))).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.ReceiptsByMigratedAtIndex(context.Background(), index)
	require.NoError(t, err)
	require.EqualValues(t, originRes.Receipts, resp)
}

func TestNodeAPI_MilestoneByIndex(t *testing.T) {
	defer gock.Off()

	var milestoneIndex uint32 = 1337
	milestoneIndexStr := strconv.Itoa(int(milestoneIndex))
	msgID := tpkg.Rand32ByteArray()

	originRes := &iotago.MilestoneResponse{
		Index:     milestoneIndex,
		MessageID: hex.EncodeToString(msgID[:]),
		Time:      time.Now().Unix(),
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRouteMilestone, milestoneIndexStr)).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.MilestoneByIndex(context.Background(), milestoneIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_MilestoneUTXOChangesByIndex(t *testing.T) {
	defer gock.Off()

	var milestoneIndex uint32 = 1337
	milestoneIndexStr := strconv.Itoa(int(milestoneIndex))

	randCreatedOutput, _ := tpkg.RandUTXOInput()
	randConsumedOutput, _ := tpkg.RandUTXOInput()

	originRes := &iotago.MilestoneUTXOChangesResponse{
		Index:           milestoneIndex,
		CreatedOutputs:  []string{randCreatedOutput.ID().ToHex()},
		ConsumedOutputs: []string{randConsumedOutput.ID().ToHex()},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRouteMilestoneUTXOChanges, milestoneIndexStr)).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.MilestoneUTXOChangesByIndex(context.Background(), milestoneIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

var sampleGossipInfo = &iotago.GossipInfo{
	Heartbeat: &iotago.GossipHeartbeat{
		SolidMilestoneIndex:  234,
		PrunedMilestoneIndex: 5872,
		LatestMilestoneIndex: 1294,
		ConnectedNeighbors:   2392,
		SyncedNeighbors:      1234,
	},
	Metrics: iotago.PeerGossipMetrics{
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

func TestNodeAPI_PeerByID(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"

	originRes := &iotago.PeerResponse{
		MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		Gossip:         sampleGossipInfo,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iotago.NodeAPIRoutePeer, peerID)).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.PeerByID(context.Background(), peerID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_RemovePeerByID(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"

	gock.New(nodeAPIUrl).
		Delete(fmt.Sprintf(iotago.NodeAPIRoutePeer, peerID)).
		Reply(200).
		Status(200)

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	err := nodeAPI.RemovePeerByID(context.Background(), peerID)
	require.NoError(t, err)
}

func TestNodeAPI_Peers(t *testing.T) {
	defer gock.Off()

	peerID1 := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
	peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

	originRes := []*iotago.PeerResponse{
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
		Get(iotago.NodeAPIRoutePeers).
		Reply(200).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.Peers(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_AddPeer(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
	multiAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)

	originRes := &iotago.PeerResponse{
		MultiAddresses: []string{multiAddr},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		Gossip:         sampleGossipInfo,
	}

	req := &iotago.AddPeerRequest{MultiAddress: multiAddr}
	gock.New(nodeAPIUrl).
		Post(iotago.NodeAPIRoutePeers).
		JSON(req).
		Reply(201).
		JSON(&iotago.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iotago.NewNodeHTTPAPIClient(nodeAPIUrl, DefZeroRentParas)
	resp, err := nodeAPI.AddPeer(context.Background(), multiAddr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}
