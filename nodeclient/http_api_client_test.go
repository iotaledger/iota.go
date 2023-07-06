// #nosec G404

package nodeclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/nodeclient"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	peerID     = "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
	nodeAPIUrl = "http://127.0.0.1:14265"
)

func nodeClient(t *testing.T) *nodeclient.Client {
	client, err := nodeclient.New(nodeAPIUrl, nodeclient.WithIOTAGoAPI(tpkg.TestAPI))
	require.NoError(t, err)
	return client
}

func TestClient_Health(t *testing.T) {
	defer gock.Off()
	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteHealth).
		Reply(200)

	nodeAPI := nodeClient(t)
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

	ts := time.Now()
	originInfo := &nodeclient.InfoResponse{
		Name:    "HORNET",
		Version: "1.0.0",
		Status: &nodeclient.InfoResNodeStatus{
			IsHealthy:            true,
			LastAcceptedBlockID:  hexutil.EncodeHex(tpkg.RandBytes(40)),
			LastConfirmedBlockID: hexutil.EncodeHex(tpkg.RandBytes(40)),
			FinalizedSlot:        iotago.SlotIndex(142857),
			ATT:                  uint64(ts.UnixNano()),
			RATT:                 uint64(ts.UnixNano()),
			CTT:                  uint64(ts.UnixNano()),
			RCTT:                 uint64(ts.UnixNano()),
			LatestCommittedSlot:  iotago.SlotIndex(142860),
			PruningSlot:          iotago.SlotIndex(142800),
		},
		BaseToken: &nodeclient.InfoResBaseToken{
			Name:            "TestCoin",
			TickerSymbol:    "TEST",
			Unit:            "TEST",
			Subunit:         "testies",
			Decimals:        6,
			UseMetricPrefix: false,
		},
		Metrics: &nodeclient.InfoResNodeMetrics{
			BlocksPerSecond:          20.0,
			ConfirmedBlocksPerSecond: 10.0,
			ConfirmedRate:            50.0,
		},
		Features: []string{"Lazers"},
	}

	var protoParams iotago.ProtocolParameters = iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions("alphanet", "atoi"),
		iotago.WithSupplyOptions(tpkg.TestTokenSupply, 500, 1, 10),
	)

	protoParamsJson, err := tpkg.TestAPI.JSONEncode(protoParams)
	require.NoError(t, err)
	protoParamsJsonRawMsg := json.RawMessage(protoParamsJson)
	originInfo.ProtocolParameters = &protoParamsJsonRawMsg

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteInfo).
		Reply(200).
		JSON(originInfo)

	nodeAPI := nodeClient(t)
	info, err := nodeAPI.Info(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originInfo, info)
	protoParams, err = originInfo.DecodeProtocolParameters()
	require.NoError(t, err)

	require.NoError(t, err)
	require.EqualValues(t, protoParams.TokenSupply(), tpkg.TestTokenSupply)
}

func TestClient_BlockIssuance(t *testing.T) {
	defer gock.Off()

	parents := []string{"733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d", "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe"}

	originRes := &nodeclient.BlockIssuanceResponse{
		StrongParents:       parents,
		WeakParents:         parents,
		ShallowLikeParents:  parents,
		LatestFinalizedSlot: iotago.SlotIndex(20),
	}

	prevID, err := iotago.SlotIdentifierFromHexString(hexutil.EncodeHex(tpkg.RandBytes(40)))
	require.NoError(t, err)
	rootsID, err := iotago.IdentifierFromHexString(hexutil.EncodeHex(tpkg.RandBytes(32)))
	require.NoError(t, err)

	commitment := &iotago.Commitment{
		Index:            iotago.SlotIndex(25),
		PrevID:           prevID,
		RootsID:          rootsID,
		CumulativeWeight: 100_000,
	}
	protoCommitmentJson, err := tpkg.TestAPI.JSONEncode(commitment)
	require.NoError(t, err)
	protoCommitmentJsonRawMsg := json.RawMessage(protoCommitmentJson)
	originRes.Commitment = &protoCommitmentJsonRawMsg

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteBlockIssuance).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.BlockIssuance(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_SubmitBlock(t *testing.T) {
	defer gock.Off()

	blockHash := tpkg.Rand40ByteArray()
	blockHashStr := hexutil.EncodeHex(blockHash[:])

	incompleteBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: &iotago.Ed25519Signature{},
		Block: &iotago.BasicBlock{
			StrongParents: tpkg.SortedRandBlockIDs(1),
		},
	}

	serializedIncompleteBlock, err := tpkg.TestAPI.Encode(incompleteBlock, serix.WithValidation())
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Post(nodeclient.RouteBlocks).
		MatchType(nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Body(bytes.NewReader(serializedIncompleteBlock)).
		Reply(200).
		AddHeader("Location", blockHashStr)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.SubmitBlock(context.Background(), incompleteBlock)
	require.NoError(t, err)
	require.EqualValues(t, blockHash, resp)
}

func TestClient_BlockMetadataByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand40ByteArray()
	parents := tpkg.SortedRandBlockIDs(1 + rand.Intn(7))

	queryHash := hexutil.EncodeHex(identifier[:])

	parentBlockIDs := make([]string, len(parents))
	for i, p := range parents {
		parentBlockIDs[i] = hexutil.EncodeHex(p[:])
	}

	originRes := &nodeclient.BlockMetadataResponse{
		BlockID:            queryHash,
		StrongParents:      parentBlockIDs,
		WeakParents:        parentBlockIDs,
		ShallowLikeParents: parentBlockIDs,
		BlockState:         "confirmed",
		TxState:            "confirmed",
		ReissuePayload:     nil,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteBlockMetadata, queryHash)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeClient(t)
	meta, err := nodeAPI.BlockMetadataByBlockID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originRes, meta)
}

func TestClient_BlockByBlockID(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand40ByteArray()
	queryHash := hexutil.EncodeHex(identifier[:])

	originBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents: tpkg.SortedRandBlockIDs(1 + rand.Intn(7)),
			Payload:       nil,
		},
	}

	data, err := tpkg.TestAPI.Encode(originBlock, serix.WithValidation())
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteBlock, queryHash)).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeClient(t)
	responseBlock, err := nodeAPI.BlockByBlockID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originBlock, responseBlock)
}

func TestClient_TransactionIncludedBlock(t *testing.T) {
	defer gock.Off()

	identifier := tpkg.Rand32ByteArray()
	queryHash := hexutil.EncodeHex(identifier[:])

	originBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents: tpkg.SortedRandBlockIDs(1 + rand.Intn(7)),
			Payload:       nil,
		},
	}

	data, err := tpkg.TestAPI.Encode(originBlock, serix.WithValidation())
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteTransactionsIncludedBlock, queryHash)).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeClient(t)
	responseBlock, err := nodeAPI.TransactionIncludedBlock(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originBlock, responseBlock)
}

func TestClient_OutputByID(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)
	data, err := tpkg.TestAPI.Encode(originOutput)
	require.NoError(t, err)

	txID := tpkg.Rand32ByteArray()

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteOutput, utxoInputId.ToHex())).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV1).
		Reply(200).
		Body(bytes.NewReader(data))

	nodeAPI := nodeClient(t)
	responseOutput, err := nodeAPI.OutputByID(context.Background(), utxoInputId)
	require.NoError(t, err)

	require.EqualValues(t, originOutput, responseOutput)
}

func TestClient_OutputMetadataByID(t *testing.T) {
	defer gock.Off()

	txID := tpkg.Rand32ByteArray()
	hexTxID := hexutil.EncodeHex(txID[:])
	originRes := &nodeclient.OutputMetadataResponse{
		BlockID:              hexutil.EncodeHex(tpkg.RandBytes(40)),
		TransactionID:        hexTxID,
		OutputIndex:          3,
		IsSpent:              true,
		CommitmentIDSpent:    hexutil.EncodeHex(tpkg.RandBytes(40)),
		TransactionIDSpent:   hexutil.EncodeHex(tpkg.RandBytes(32)),
		IncludedCommitmentID: hexutil.EncodeHex(tpkg.RandBytes(40)),
		LatestCommitmentID:   hexutil.EncodeHex(tpkg.RandBytes(40)),
	}

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteOutputMetadata, utxoInputId.ToHex())).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.OutputMetadataByID(context.Background(), utxoInputId)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	resTxID, err := resp.TxID()
	require.NoError(t, err)
	require.EqualValues(t, txID, *resTxID)
}

func TestClient_CommitmentByID(t *testing.T) {
	defer gock.Off()

	var slotIndex iotago.SlotIndex = 5

	commitmentID := iotago.NewSlotIdentifier(slotIndex, tpkg.Rand32ByteArray())
	commitment := iotago.NewCommitment(tpkg.TestAPI.Version(), slotIndex, iotago.NewSlotIdentifier(slotIndex-1, tpkg.Rand32ByteArray()), tpkg.Rand32ByteArray(), tpkg.RandUint64(math.MaxUint64))

	originRes := &nodeclient.CommitmentDetailsResponse{
		Index:            commitment.Index,
		PrevID:           commitment.PrevID.ToHex(),
		RootsID:          commitment.RootsID.ToHex(),
		CumulativeWeight: commitment.CumulativeWeight,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteCommitmentByID, commitmentID.ToHex())).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentByID(context.Background(), commitmentID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_CommitmentUTXOChangesByID(t *testing.T) {
	defer gock.Off()

	commitmentID := iotago.NewSlotIdentifier(5, tpkg.Rand32ByteArray())

	randCreatedOutput := tpkg.RandUTXOInput()
	randConsumedOutput := tpkg.RandUTXOInput()

	originRes := &nodeclient.UTXOChangesResponse{
		Index:           1337,
		CreatedOutputs:  []string{randCreatedOutput.ID().ToHex()},
		ConsumedOutputs: []string{randConsumedOutput.ID().ToHex()},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteCommitmentByIDUTXOChanges, commitmentID.ToHex())).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentUTXOChangesByID(context.Background(), commitmentID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_CommitmentByIndex(t *testing.T) {
	defer gock.Off()

	var slotIndex iotago.SlotIndex = 1337

	commitment := iotago.NewCommitment(tpkg.TestAPI.Version(), slotIndex, iotago.NewSlotIdentifier(slotIndex-1, tpkg.Rand32ByteArray()), tpkg.Rand32ByteArray(), tpkg.RandUint64(math.MaxUint64))

	originRes := &nodeclient.CommitmentDetailsResponse{
		Index:            commitment.Index,
		PrevID:           commitment.PrevID.ToHex(),
		RootsID:          commitment.RootsID.ToHex(),
		CumulativeWeight: commitment.CumulativeWeight,
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteCommitmentByIndex, slotIndex)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentByIndex(context.Background(), slotIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_CommitmentUTXOChangesByIndex(t *testing.T) {
	defer gock.Off()

	var slotIndex iotago.SlotIndex = 1337

	randCreatedOutput := tpkg.RandUTXOInput()
	randConsumedOutput := tpkg.RandUTXOInput()

	originRes := &nodeclient.UTXOChangesResponse{
		Index:           slotIndex,
		CreatedOutputs:  []string{randCreatedOutput.ID().ToHex()},
		ConsumedOutputs: []string{randConsumedOutput.ID().ToHex()},
	}

	gock.New(nodeAPIUrl).
		Get(fmt.Sprintf(nodeclient.RouteCommitmentByIndexUTXOChanges, slotIndex)).
		Reply(200).
		JSON(originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentUTXOChangesByIndex(context.Background(), slotIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

var sampleGossipInfo = &nodeclient.GossipInfo{
	Heartbeat: &nodeclient.GossipHeartbeat{
		SolidMilestoneIndex:  234,
		PrunedMilestoneIndex: 5872,
		LatestMilestoneIndex: 1294,
		ConnectedPeers:       2392,
		SyncedPeers:          1234,
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

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.PeerByID(context.Background(), peerID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_RemovePeerByID(t *testing.T) {
	defer gock.Off()

	gock.New(nodeAPIUrl).
		Delete(fmt.Sprintf(nodeclient.RoutePeer, peerID)).
		Reply(200).
		Status(200)

	nodeAPI := nodeClient(t)
	err := nodeAPI.RemovePeerByID(context.Background(), peerID)
	require.NoError(t, err)
}

func TestClient_Peers(t *testing.T) {
	defer gock.Off()

	peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

	originRes := []*nodeclient.PeerResponse{
		{
			MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
			ID:             peerID,
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

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.Peers(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_AddPeer(t *testing.T) {
	defer gock.Off()

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

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.AddPeer(context.Background(), multiAddr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}
