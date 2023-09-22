// #nosec G404
//
//nolint:dupl
package nodeclient_test

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/nodeclient"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	peerID     = "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
	nodeAPIUrl = "http://127.0.0.1:14265"
)

var (
	protoParams = iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions("alphanet", "atoi"),
		iotago.WithSupplyOptions(tpkg.TestTokenSupply, 500, 1, 10, 100, 100, 100),
	)

	mockAPI = iotago.V3API(protoParams)
)

//nolint:unparam // false positive
func mockGetJSON(route string, status int, body interface{}, persist ...bool) {
	m := gock.New(nodeAPIUrl).
		Get(route)

	if len(persist) > 0 && persist[0] {
		m.Persist()
	}

	m.Reply(status).SetHeader("Content-Type", nodeclient.MIMEApplicationJSON).
		BodyString(string(lo.PanicOnErr(mockAPI.JSONEncode(body))))
}

//nolint:unparam // false positive
func mockGetJSONWithParams(route string, status int, body interface{}, params map[string]string, persist ...bool) {
	m := gock.New(nodeAPIUrl).
		Get(route).
		MatchParams(params)

	if len(persist) > 0 && persist[0] {
		m.Persist()
	}

	m.Reply(status).
		SetHeader("Content-Type", nodeclient.MIMEApplicationJSON).
		BodyString(string(lo.PanicOnErr(mockAPI.JSONEncode(body))))
}

//nolint:unparam // false positive
func mockPostJSON(route string, status int, req interface{}, resp interface{}) {
	gock.New(nodeAPIUrl).
		Post(route).
		MatchHeader("Content-Type", nodeclient.MIMEApplicationJSON).
		BodyString(string(lo.PanicOnErr(mockAPI.JSONEncode(req)))).
		Reply(status).
		SetHeader("Content-Type", nodeclient.MIMEApplicationJSON).
		BodyString(string(lo.PanicOnErr(mockAPI.JSONEncode(resp))))
}

//nolint:unparam // false positive
func mockGetBinary(route string, status int, body interface{}, persist ...bool) {
	m := gock.New(nodeAPIUrl).
		Get(route).
		MatchHeader("Accept", nodeclient.MIMEApplicationVendorIOTASerializerV2)

	if len(persist) > 0 && persist[0] {
		m.Persist()
	}

	m.Reply(status).
		SetHeader("Content-Type", nodeclient.MIMEApplicationVendorIOTASerializerV2).
		BodyString(string(lo.PanicOnErr(mockAPI.Encode(body))))
}

//nolint:thelper
func nodeClient(t *testing.T) *nodeclient.Client {

	ts := time.Now()
	originInfo := &apimodels.InfoResponse{
		Name:    "HORNET",
		Version: "1.0.0",
		Status: &apimodels.InfoResNodeStatus{
			IsHealthy:                   true,
			LatestAcceptedBlockSlot:     tpkg.RandSlotIndex(),
			LatestConfirmedBlockSlot:    tpkg.RandSlotIndex(),
			LatestFinalizedSlot:         iotago.SlotIndex(142857),
			AcceptedTangleTime:          ts,
			RelativeAcceptedTangleTime:  ts,
			ConfirmedTangleTime:         ts,
			RelativeConfirmedTangleTime: ts,
			LatestCommitmentID:          tpkg.Rand40ByteArray(),
			PruningEpoch:                iotago.EpochIndex(142800),
		},
		ProtocolParameters: []*apimodels.InfoResProtocolParameters{
			{
				StartEpoch: 0,
				Parameters: protoParams,
			},
		},
		BaseToken: &apimodels.InfoResBaseToken{
			Name:            "TestCoin",
			TickerSymbol:    "TEST",
			Unit:            "TEST",
			Subunit:         "testies",
			Decimals:        6,
			UseMetricPrefix: false,
		},
		Metrics: &apimodels.InfoResNodeMetrics{
			BlocksPerSecond:          20.0,
			ConfirmedBlocksPerSecond: 10.0,
			ConfirmationRate:         50.0,
		},
		Features: []string{"Lazers"},
	}

	mockGetJSON(nodeclient.RouteInfo, 200, originInfo)

	client, err := nodeclient.New(nodeAPIUrl)
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

func TestClient_BlockIssuance(t *testing.T) {
	defer gock.Off()

	parentsHex := []string{"0x733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d0000000000000000", "0x5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe0000000000000000"}
	parents, err := iotago.BlockIDsFromHexString(parentsHex)
	require.NoError(t, err)

	originRes := &apimodels.IssuanceBlockHeaderResponse{
		StrongParents:       parents,
		WeakParents:         parents,
		ShallowLikeParents:  parents,
		LatestFinalizedSlot: iotago.SlotIndex(20),
	}

	prevID, err := iotago.SlotIdentifierFromHexString(hexutil.EncodeHex(tpkg.RandBytes(40)))
	require.NoError(t, err)
	rootsID, err := iotago.IdentifierFromHexString(hexutil.EncodeHex(tpkg.RandBytes(32)))
	require.NoError(t, err)

	originRes.Commitment = &iotago.Commitment{
		ProtocolVersion:      1,
		Index:                iotago.SlotIndex(25),
		PreviousCommitmentID: prevID,
		RootsID:              rootsID,
		CumulativeWeight:     100_000,
	}

	mockGetJSON(nodeclient.RouteBlockIssuance, 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.BlockIssuance(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Congestion(t *testing.T) {
	defer gock.Off()

	accID := tpkg.RandAccountID()

	originRes := &apimodels.CongestionResponse{
		SlotIndex:            iotago.SlotIndex(20),
		Ready:                true,
		ReferenceManaCost:    iotago.Mana(1000),
		BlockIssuanceCredits: iotago.BlockIssuanceCredits(1000),
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteCongestion, accID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.Congestion(context.Background(), accID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Rewards(t *testing.T) {
	defer gock.Off()

	outID := tpkg.RandOutputID(1)

	originRes := &apimodels.ManaRewardsResponse{
		EpochStart: iotago.EpochIndex(20),
		EpochEnd:   iotago.EpochIndex(30),
		Rewards:    iotago.Mana(1000),
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteRewards, outID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.Rewards(context.Background(), outID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Validators(t *testing.T) {
	defer gock.Off()

	originRes := &apimodels.ValidatorsResponse{Validators: []*apimodels.ValidatorResponse{
		{
			AccountID:                      tpkg.RandAccountID(),
			StakingEpochEnd:                iotago.EpochIndex(123),
			PoolStake:                      iotago.BaseToken(100),
			ValidatorStake:                 iotago.BaseToken(10),
			FixedCost:                      iotago.Mana(10),
			Active:                         true,
			LatestSupportedProtocolVersion: 1,
		},
		{
			AccountID:                      tpkg.RandAccountID(),
			StakingEpochEnd:                iotago.EpochIndex(124),
			PoolStake:                      iotago.BaseToken(1000),
			ValidatorStake:                 iotago.BaseToken(100),
			FixedCost:                      iotago.Mana(20),
			Active:                         true,
			LatestSupportedProtocolVersion: 1,
		},
	}}

	mockGetJSON(nodeclient.RouteValidators, 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.Validators(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_StakingByAccountID(t *testing.T) {
	defer gock.Off()

	accID := tpkg.RandAccountID()
	originRes := &apimodels.ValidatorResponse{
		AccountID:                      accID,
		StakingEpochEnd:                iotago.EpochIndex(123),
		PoolStake:                      iotago.BaseToken(100),
		ValidatorStake:                 iotago.BaseToken(10),
		FixedCost:                      iotago.Mana(10),
		Active:                         true,
		LatestSupportedProtocolVersion: 1,
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteValidatorsAccount, accID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.StakingAccount(context.Background(), accID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Committee(t *testing.T) {
	defer gock.Off()

	originRes := &apimodels.CommitteeResponse{
		EpochIndex:          iotago.EpochIndex(123),
		TotalStake:          1000_1000,
		TotalValidatorStake: 100_000,
		Committee: []*apimodels.CommitteeMemberResponse{
			{
				AccountID:      tpkg.RandAccountID(),
				PoolStake:      1000_000,
				ValidatorStake: 100_000,
				FixedCost:      iotago.Mana(100),
			},
		},
	}

	mockGetJSON(nodeclient.RouteCommittee, 200, originRes)
	nodeAPI := nodeClient(t)
	res, err := nodeAPI.Committee(context.Background())
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
			StrongParents:      tpkg.SortedRandBlockIDs(1),
			WeakParents:        iotago.BlockIDs{},
			ShallowLikeParents: iotago.BlockIDs{},
		},
	}

	serializedIncompleteBlock, err := tpkg.TestAPI.Encode(incompleteBlock, serix.WithValidation())
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Post(nodeclient.RouteBlocks).
		MatchType(nodeclient.MIMEApplicationVendorIOTASerializerV2).
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

	identifier := tpkg.RandBlockID()

	originRes := &apimodels.BlockMetadataResponse{
		BlockID:    identifier,
		BlockState: apimodels.BlockStateConfirmed.String(),
		TxState:    apimodels.TransactionStateConfirmed.String(),
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteBlockMetadata, identifier.ToHex()), 200, originRes)

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
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents:      tpkg.SortedRandBlockIDs(1 + rand.Intn(7)),
			WeakParents:        iotago.BlockIDs{},
			ShallowLikeParents: iotago.BlockIDs{},
			Payload:            nil,
		},
	}

	mockGetBinary(fmt.Sprintf(nodeclient.RouteBlock, queryHash), 200, originBlock)

	nodeAPI := nodeClient(t)
	responseBlock, err := nodeAPI.BlockByBlockID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originBlock, responseBlock)
}

func TestClient_TransactionIncludedBlock(t *testing.T) {
	defer gock.Off()

	txID := tpkg.Rand40ByteArray()
	queryHash := hexutil.EncodeHex(txID[:])

	originBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents:      tpkg.SortedRandBlockIDs(1 + rand.Intn(7)),
			WeakParents:        iotago.BlockIDs{},
			ShallowLikeParents: iotago.BlockIDs{},
			Payload:            nil,
		},
	}

	mockGetBinary(fmt.Sprintf(nodeclient.RouteTransactionsIncludedBlock, queryHash), 200, originBlock)

	nodeAPI := nodeClient(t)
	responseBlock, err := nodeAPI.TransactionIncludedBlock(context.Background(), txID)
	require.NoError(t, err)
	require.EqualValues(t, originBlock, responseBlock)
}

func TestClient_OutputByID(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)

	txID := tpkg.Rand40ByteArray()

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputID := utxoInput.OutputID()

	mockGetBinary(fmt.Sprintf(nodeclient.RouteOutput, utxoInputID.ToHex()), 200, originOutput)

	nodeAPI := nodeClient(t)
	responseOutput, err := nodeAPI.OutputByID(context.Background(), utxoInputID)
	require.NoError(t, err)

	require.EqualValues(t, originOutput, responseOutput)
}

func TestClient_OutputMetadataByID(t *testing.T) {
	defer gock.Off()

	txID := tpkg.Rand40ByteArray()
	originRes := &apimodels.OutputMetadataResponse{
		BlockID:              tpkg.RandBlockID(),
		TransactionID:        txID,
		OutputIndex:          3,
		IsSpent:              true,
		CommitmentIDSpent:    tpkg.Rand40ByteArray(),
		TransactionIDSpent:   tpkg.Rand40ByteArray(),
		IncludedCommitmentID: tpkg.Rand40ByteArray(),
		LatestCommitmentID:   tpkg.Rand40ByteArray(),
	}

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputID := utxoInput.OutputID()

	mockGetJSON(fmt.Sprintf(nodeclient.RouteOutputMetadata, utxoInputID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.OutputMetadataByID(context.Background(), utxoInputID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	require.EqualValues(t, txID, resp.TransactionID)
}

func TestClient_CommitmentByID(t *testing.T) {
	defer gock.Off()

	var slotIndex iotago.SlotIndex = 5

	commitmentID := iotago.NewSlotIdentifier(slotIndex, tpkg.Rand32ByteArray())
	commitment := iotago.NewCommitment(tpkg.TestAPI.Version(), slotIndex, iotago.NewSlotIdentifier(slotIndex-1, tpkg.Rand32ByteArray()), tpkg.Rand32ByteArray(), tpkg.RandUint64(math.MaxUint64), tpkg.RandMana(math.MaxUint64))

	originRes := &iotago.Commitment{
		Index:                commitment.Index,
		PreviousCommitmentID: commitment.PreviousCommitmentID,
		RootsID:              commitment.RootsID,
		CumulativeWeight:     commitment.CumulativeWeight,
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteCommitmentByID, commitmentID.ToHex()), 200, originRes)

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

	originRes := &apimodels.UTXOChangesResponse{
		Index: 1337,
		CreatedOutputs: iotago.OutputIDs{
			randCreatedOutput.OutputID(),
		},
		ConsumedOutputs: iotago.OutputIDs{
			randConsumedOutput.OutputID(),
		},
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteCommitmentByIDUTXOChanges, commitmentID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentUTXOChangesByID(context.Background(), commitmentID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_CommitmentByIndex(t *testing.T) {
	defer gock.Off()

	var slotIndex iotago.SlotIndex = 1337

	commitment := iotago.NewCommitment(tpkg.TestAPI.Version(), slotIndex, iotago.NewSlotIdentifier(slotIndex-1, tpkg.Rand32ByteArray()), tpkg.Rand32ByteArray(), tpkg.RandUint64(math.MaxUint64), tpkg.RandMana(math.MaxUint64))

	originRes := &iotago.Commitment{
		Index:                commitment.Index,
		PreviousCommitmentID: commitment.PreviousCommitmentID,
		RootsID:              commitment.RootsID,
		CumulativeWeight:     commitment.CumulativeWeight,
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteCommitmentByIndex, slotIndex), 200, originRes)

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

	originRes := &apimodels.UTXOChangesResponse{
		Index: slotIndex,
		CreatedOutputs: iotago.OutputIDs{
			randCreatedOutput.OutputID(),
		},
		ConsumedOutputs: iotago.OutputIDs{
			randConsumedOutput.OutputID(),
		},
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RouteCommitmentByIndexUTXOChanges, slotIndex), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentUTXOChangesByIndex(context.Background(), slotIndex)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

var sampleGossipInfo = &apimodels.GossipInfo{
	Heartbeat: &apimodels.GossipHeartbeat{
		SolidSlotIndex:  234,
		PrunedSlotIndex: 5872,
		LatestSlotIndex: 1294,
		ConnectedPeers:  2392,
		SyncedPeers:     1234,
	},
	Metrics: &apimodels.PeerGossipMetrics{
		NewBlocks:             40,
		KnownBlocks:           60,
		ReceivedBlocks:        100,
		ReceivedBlockRequests: 345,
		ReceivedSlotRequests:  194,
		ReceivedHeartbeats:    5,
		SentBlocks:            492,
		SentBlockRequests:     2396,
		SentSlotRequests:      9837,
		SentHeartbeats:        3,
		DroppedPackets:        10,
	},
}

func TestClient_PeerByID(t *testing.T) {
	defer gock.Off()

	originRes := &apimodels.PeerInfo{
		MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		Gossip:         sampleGossipInfo,
	}

	mockGetJSON(fmt.Sprintf(nodeclient.RoutePeer, peerID), 200, originRes)

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

	originRes := &apimodels.PeersResponse{
		Peers: []*apimodels.PeerInfo{
			{
				ID:             peerID,
				MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
				Relation:       "autopeered",
				Gossip:         sampleGossipInfo,
				Connected:      true,
			},
			{
				ID:             peerID2,
				MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID2)},
				Alias:          "Peer2",
				Relation:       "static",
				Gossip:         sampleGossipInfo,
				Connected:      true,
			},
		},
	}

	mockGetJSON(nodeclient.RoutePeers, 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.Peers(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_AddPeer(t *testing.T) {
	defer gock.Off()

	multiAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)

	originRes := &apimodels.PeerInfo{
		ID:             peerID,
		MultiAddresses: []string{multiAddr},
		Relation:       "autopeered",
		Connected:      true,
		Gossip:         sampleGossipInfo,
	}

	req := &apimodels.AddPeerRequest{MultiAddress: multiAddr}

	mockPostJSON(nodeclient.RoutePeers, 201, req, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.AddPeer(context.Background(), multiAddr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}
