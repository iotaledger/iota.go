package iota_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const nodeAPIUrl = "http://127.0.0.1:14265"

func TestNodeAPI_Health(t *testing.T) {
	defer gock.Off()
	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteHealth).
		Reply(200)

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	healthy, err := nodeAPI.Health()
	require.NoError(t, err)
	require.True(t, healthy)

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteHealth).
		Reply(503)

	healthy, err = nodeAPI.Health()
	require.NoError(t, err)
	require.False(t, healthy)
}

func TestNodeAPI_Info(t *testing.T) {
	defer gock.Off()

	originInfo := &iota.NodeInfoResponse{
		Name:                     "HORNET",
		Version:                  "1.0.0",
		IsHealthy:                true,
		CoordinatorPublicKey:     "733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		LatestMilestoneMessageID: "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe",
		LatestMilestoneIndex:     1337,
		SolidMilestoneMessageID:  "598f7a3186bf7291b8199a3147bb2a81d19b89ac545788b4e5d8adbee7db0f13",
		SolidMilestoneIndex:      666,
		PruningIndex:             142857,
		Features:                 []string{"Lazers"},
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteInfo).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originInfo})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	info, err := nodeAPI.Info()
	require.NoError(t, err)
	require.EqualValues(t, originInfo, info)
}

func TestNodeAPI_Tips(t *testing.T) {
	defer gock.Off()

	originRes := &iota.NodeTipsResponse{
		Tip1: "733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		Tip2: "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe",
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteTips).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	tips, err := nodeAPI.Tips()
	require.NoError(t, err)
	require.EqualValues(t, originRes, tips)
}

func TestNodeAPI_SubmitMessage(t *testing.T) {
	defer gock.Off()

	msgHash := rand32ByteHash()
	msgHashStr := hex.EncodeToString(msgHash[:])

	incompleteMsg := &iota.Message{Version: 1}
	completeMsg := &iota.Message{
		Version: 1,
		Parent1: rand32ByteHash(),
		Parent2: rand32ByteHash(),
		Payload: nil,
		Nonce:   3495721389537486,
	}

	serializedCompleteMsg, err := completeMsg.Serialize(iota.DeSeriModeNoValidation)
	require.NoError(t, err)

	// we need to do this, otherwise gock doesn't match the body
	gock.BodyTypes = append(gock.BodyTypes, "application/octet-stream")
	gock.BodyTypeAliases["octet"] = "application/octet-stream"

	serializedIncompleteMsg, err := incompleteMsg.Serialize(iota.DeSeriModePerformValidation)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Post(iota.NodeAPIRouteMessages).
		MatchType("octet").
		Body(bytes.NewReader(serializedIncompleteMsg)).
		Reply(200).
		AddHeader("Location", msgHashStr)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iota.NodeAPIRouteMessageBytes, msgHashStr)).
		Reply(200).
		Body(bytes.NewReader(serializedCompleteMsg))

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.SubmitMessage(incompleteMsg)
	require.NoError(t, err)
	require.EqualValues(t, completeMsg, resp)
}

func TestNodeAPI_MessageIDsByIndex(t *testing.T) {
	defer gock.Off()
	index := "बेकार पाठ"

	id1 := rand32ByteHash()
	id2 := rand32ByteHash()
	id3 := rand32ByteHash()

	msgIDsByIndex := &iota.MessageIDsByIndexResponse{
		Index:      index,
		MaxResults: 1000,
		Count:      3,
		MessageIDs: []string{
			hex.EncodeToString(id1[:]),
			hex.EncodeToString(id2[:]),
			hex.EncodeToString(id3[:]),
		},
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteMessages).
		MatchParam("index", index).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: msgIDsByIndex})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resMsgIDsByIndex, err := nodeAPI.MessageIDsByIndex(index)
	require.NoError(t, err)
	require.EqualValues(t, msgIDsByIndex, resMsgIDsByIndex)
}

func TestNodeAPI_MessageMetadataByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := rand32ByteHash()
	parent1 := rand32ByteHash()
	parent2 := rand32ByteHash()

	queryHash := hex.EncodeToString(identifier[:])
	parent1MessageID := hex.EncodeToString(parent1[:])
	parent2MessageID := hex.EncodeToString(parent2[:])

	originRes := &iota.MessageMetadataResponse{
		MessageID:                  queryHash,
		Parent1:                    parent1MessageID,
		Parent2:                    parent2MessageID,
		Solid:                      true,
		ReferencedByMilestoneIndex: nil,
		LedgerInclusionState:       nil,
		ShouldPromote:              nil,
		ShouldReattach:             nil,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iota.NodeAPIRouteMessageMetadata, queryHash)).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	meta, err := nodeAPI.MessageMetadataByMessageID(identifier)
	require.NoError(t, err)
	require.EqualValues(t, originRes, meta)
}

func TestNodeAPI_MessageByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := rand32ByteHash()
	queryHash := hex.EncodeToString(identifier[:])

	originMsg := &iota.Message{
		Version: 1,
		Parent1: rand32ByteHash(),
		Parent2: rand32ByteHash(),
		Payload: nil,
		Nonce:   16345984576234,
	}

	data, err := originMsg.Serialize(iota.DeSeriModePerformValidation)
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iota.NodeAPIRouteMessageBytes, queryHash)).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	responseMsg, err := nodeAPI.MessageByMessageID(identifier)
	require.NoError(t, err)
	require.EqualValues(t, originMsg, responseMsg)
}

func TestNodeAPI_ChildrenByMessageID(t *testing.T) {
	defer gock.Off()

	msgID := rand32ByteHash()
	hexMsgID := hex.EncodeToString(msgID[:])

	child1 := rand32ByteHash()
	child2 := rand32ByteHash()
	child3 := rand32ByteHash()

	originRes := &iota.ChildrenResponse{
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
		Get(fmt.Sprintf(iota.NodeAPIRouteMessageChildren, hexMsgID)).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	res, err := nodeAPI.ChildrenByMessageID(msgID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestNodeAPI_OutputByID(t *testing.T) {
	defer gock.Off()

	originOutput, _ := randSigLockedSingleOutput(iota.AddressEd25519)
	sigDepJson, err := originOutput.MarshalJSON()
	require.NoError(t, err)
	rawMsgSigDepJson := json.RawMessage(sigDepJson)

	txID := rand32ByteHash()
	hexTxID := hex.EncodeToString(txID[:])
	originRes := &iota.NodeOutputResponse{
		TransactionID: hexTxID,
		OutputIndex:   3,
		Spent:         true,
		RawOutput:     &rawMsgSigDepJson,
	}

	utxoInput := &iota.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iota.NodeAPIRouteOutput, utxoInputId.ToHex())).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.OutputByID(utxoInputId)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	resTxID, err := resp.TxID()
	require.NoError(t, err)
	require.EqualValues(t, txID, *resTxID)
}

func TestNodeAPI_BalanceByAddress(t *testing.T) {
	defer gock.Off()

	ed25519Addr, _ := randEd25519Addr()
	ed25519AddrHex := ed25519Addr.String()

	originRes := &iota.AddressBalanceResponse{
		Address:    ed25519AddrHex,
		MaxResults: 1000,
		Count:      1337,
		Balance:    13371337,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iota.NodeAPIRouteAddressBalance, ed25519AddrHex)).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.BalanceByAddress(ed25519AddrHex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_OutputIDsByAddress(t *testing.T) {
	defer gock.Off()

	ed25519Addr, _ := randEd25519Addr()
	ed25519AddrHex := ed25519Addr.String()

	output1 := rand32ByteHash()
	output2 := rand32ByteHash()
	output3 := rand32ByteHash()
	originRes := &iota.AddressOutputsResponse{
		Address:    ed25519AddrHex,
		MaxResults: 1000,
		Count:      2,
		OutputIDs: []string{
			hex.EncodeToString(output1[:]),
			hex.EncodeToString(output2[:]),
		},
	}

	originResWithUnspent := &iota.AddressOutputsResponse{
		Address:    ed25519AddrHex,
		MaxResults: 1000,
		Count:      3,
		OutputIDs: []string{
			hex.EncodeToString(output1[:]),
			hex.EncodeToString(output2[:]),
			hex.EncodeToString(output3[:]),
		},
	}

	route := fmt.Sprintf(iota.NodeAPIRouteAddressOutputs, ed25519AddrHex)
	gock.New(nodeAPIUrl).
		Get(route).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.OutputIDsByAddress(ed25519AddrHex, false)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	gock.New(nodeAPIUrl).
		Get(route).
		MatchParam("include-spent", "true").
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originResWithUnspent})

	resp, err = nodeAPI.OutputIDsByAddress(ed25519AddrHex, true)
	require.NoError(t, err)
	require.EqualValues(t, originResWithUnspent, resp)
}

func TestNodeAPI_MilestoneByIndex(t *testing.T) {
	defer gock.Off()

	var milestoneIndex uint32 = 1337
	milestoneIndexStr := strconv.Itoa(int(milestoneIndex))
	msgID := rand32ByteHash()

	originRes := &iota.MilestoneResponse{
		Index:     milestoneIndex,
		MessageID: hex.EncodeToString(msgID[:]),
		Time:      time.Now().Unix(),
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iota.NodeAPIRouteMilestone, milestoneIndexStr)).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.MilestoneByIndex(milestoneIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_PeerByID(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"

	originRes := &iota.PeerResponse{
		MultiAddress: fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID),
		ID:           peerID,
		Connected:    true,
		Relation:     "autopeered",
		GossipMetrics: &iota.PeerGossipMetrics{
			SentPackets:        100,
			DroppedSentPackets: 10,
			ReceivedHeartbeats: 5,
			SentHeartbeats:     3,
			ReceivedMessages:   100,
			NewMessages:        40,
			KnownMessages:      60,
		},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(iota.NodeAPIRoutePeer, peerID)).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.PeerByID(peerID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_RemovePeerByID(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"

	gock.New(nodeAPIUrl).
		Delete(fmt.Sprintf(iota.NodeAPIRoutePeer, peerID)).
		Reply(200).
		Status(200)

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	err := nodeAPI.RemovePeerByID(peerID)
	require.NoError(t, err)
}

func TestNodeAPI_Peers(t *testing.T) {
	defer gock.Off()

	peerID1 := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
	peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

	originRes := []*iota.PeerResponse{
		{
			MultiAddress: fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID1),
			ID:           peerID1,
			Connected:    true,
			Relation:     "autopeered",
			GossipMetrics: &iota.PeerGossipMetrics{
				SentPackets:        100,
				DroppedSentPackets: 10,
				ReceivedHeartbeats: 5,
				SentHeartbeats:     3,
				ReceivedMessages:   100,
				NewMessages:        40,
				KnownMessages:      60,
			},
		},
		{
			MultiAddress: fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID2),
			ID:           peerID2,
			Connected:    true,
			Relation:     "static",
			GossipMetrics: &iota.PeerGossipMetrics{
				SentPackets:        100,
				DroppedSentPackets: 10,
				ReceivedHeartbeats: 5,
				SentHeartbeats:     3,
				ReceivedMessages:   100,
				NewMessages:        40,
				KnownMessages:      60,
			},
		},
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRoutePeers).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.Peers()
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestNodeAPI_AddPeer(t *testing.T) {
	defer gock.Off()

	peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"

	originRes := &iota.PeerResponse{
		MultiAddress: fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID),
		ID:           peerID,
		Connected:    true,
		Relation:     "autopeered",
		GossipMetrics: &iota.PeerGossipMetrics{
			SentPackets:        100,
			DroppedSentPackets: 10,
			ReceivedHeartbeats: 5,
			SentHeartbeats:     3,
			ReceivedMessages:   100,
			NewMessages:        40,
			KnownMessages:      60,
		},
	}

	req := &iota.AddPeerRequest{
		MultiAddress: fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID),
	}

	gock.New(nodeAPIUrl).
		Post(iota.NodeAPIRoutePeers).
		MatchType("json").
		JSON(req).
		Reply(201).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.AddPeer(peerID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}
