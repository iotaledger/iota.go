// #nosec G404
//
//nolint:dupl,gosec,forcetypeassert
package nodeclient_test

import (
	"bytes"
	"context"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/nodeclient"
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

	m.Reply(status).SetHeader("Content-Type", api.MIMEApplicationJSON).
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
		SetHeader("Content-Type", api.MIMEApplicationJSON).
		BodyString(string(lo.PanicOnErr(mockAPI.JSONEncode(body))))
}

//nolint:unparam // false positive
func mockPostJSON(route string, status int, req interface{}, resp interface{}) {
	gock.New(nodeAPIUrl).
		Post(route).
		MatchHeader("Content-Type", api.MIMEApplicationJSON).
		BodyString(string(lo.PanicOnErr(mockAPI.JSONEncode(req)))).
		Reply(status).
		SetHeader("Content-Type", api.MIMEApplicationJSON).
		BodyString(string(lo.PanicOnErr(mockAPI.JSONEncode(resp))))
}

//nolint:unparam // false positive
func mockGetBinary(route string, status int, body interface{}, persist ...bool) {
	m := gock.New(nodeAPIUrl).
		Get(route).
		MatchHeader("Accept", api.MIMEApplicationVendorIOTASerializerV2)

	if len(persist) > 0 && persist[0] {
		m.Persist()
	}

	m.Reply(status).
		SetHeader("Content-Type", api.MIMEApplicationVendorIOTASerializerV2).
		BodyString(string(lo.PanicOnErr(mockAPI.Encode(body))))
}

//nolint:thelper
func nodeClient(t *testing.T) *nodeclient.Client {

	ts := time.Now()
	originInfo := &api.InfoResponse{
		Name:    "HORNET",
		Version: "1.0.0",
		Status: &api.InfoResNodeStatus{
			IsHealthy:                   true,
			LatestAcceptedBlockSlot:     tpkg.RandSlot(),
			LatestConfirmedBlockSlot:    tpkg.RandSlot(),
			LatestFinalizedSlot:         iotago.SlotIndex(142857),
			AcceptedTangleTime:          ts,
			RelativeAcceptedTangleTime:  ts,
			ConfirmedTangleTime:         ts,
			RelativeConfirmedTangleTime: ts,
			LatestCommitmentID:          tpkg.Rand36ByteArray(),
			PruningEpoch:                iotago.EpochIndex(142800),
		},
		ProtocolParameters: []*api.InfoResProtocolParameters{
			{
				StartEpoch: 0,
				Parameters: protoParams,
			},
		},
		BaseToken: &api.InfoResBaseToken{
			Name:         "TestCoin",
			TickerSymbol: "TEST",
			Unit:         "TEST",
			Subunit:      "testies",
			Decimals:     6,
		},
		Metrics: &api.InfoResNodeMetrics{
			BlocksPerSecond:          20.0,
			ConfirmedBlocksPerSecond: 10.0,
			ConfirmationRate:         50.0,
		},
		Features: []string{"Lazers"},
	}

	mockGetJSON(api.CoreRouteInfo, 200, originInfo)

	client, err := nodeclient.New(nodeAPIUrl)
	require.NoError(t, err)

	return client
}

func TestClient_Health(t *testing.T) {
	defer gock.Off()

	gock.New(nodeAPIUrl).
		Get(api.RouteHealth).
		Reply(200)

	nodeAPI := nodeClient(t)
	healthy, err := nodeAPI.Health(context.Background())
	require.NoError(t, err)
	require.True(t, healthy)

	gock.New(nodeAPIUrl).
		Get(api.RouteHealth).
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

	originRes := &api.IssuanceBlockHeaderResponse{
		StrongParents:       parents,
		WeakParents:         parents,
		ShallowLikeParents:  parents,
		LatestFinalizedSlot: iotago.SlotIndex(20),
	}

	prevID, err := iotago.CommitmentIDFromHexString(hexutil.EncodeHex(tpkg.RandBytes(40)))
	require.NoError(t, err)
	rootsID, err := iotago.IdentifierFromHexString(hexutil.EncodeHex(tpkg.RandBytes(32)))
	require.NoError(t, err)

	originRes.Commitment = &iotago.Commitment{
		ProtocolVersion:      1,
		Slot:                 iotago.SlotIndex(25),
		PreviousCommitmentID: prevID,
		RootsID:              rootsID,
		CumulativeWeight:     100_000,
	}

	mockGetJSON(api.CoreRouteBlockIssuance, 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.BlockIssuance(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Congestion(t *testing.T) {
	defer gock.Off()

	accountAddress := tpkg.RandAccountID().ToAddress().(*iotago.AccountAddress)

	originRes := &api.CongestionResponse{
		Slot:                 iotago.SlotIndex(20),
		Ready:                true,
		ReferenceManaCost:    iotago.Mana(1000),
		BlockIssuanceCredits: iotago.BlockIssuanceCredits(1000),
	}

	nodeAPI := nodeClient(t)
	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteCongestion, api.ParameterBech32Address, accountAddress.Bech32(nodeAPI.CommittedAPI().ProtocolParameters().Bech32HRP())), 200, originRes)

	res, err := nodeAPI.Congestion(context.Background(), accountAddress)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Rewards(t *testing.T) {
	defer gock.Off()

	outID := tpkg.RandOutputID(1)

	originRes := &api.ManaRewardsResponse{
		EpochStart: iotago.EpochIndex(20),
		EpochEnd:   iotago.EpochIndex(30),
		Rewards:    iotago.Mana(1000),
	}

	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteRewards, api.ParameterOutputID, outID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.Rewards(context.Background(), outID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Validators(t *testing.T) {
	defer gock.Off()

	originRes := &api.ValidatorsResponse{Validators: []*api.ValidatorResponse{
		{
			AddressBech32:                  tpkg.RandAccountID().ToAddress().Bech32(iotago.PrefixTestnet),
			StakingEpochEnd:                iotago.EpochIndex(123),
			PoolStake:                      iotago.BaseToken(100),
			ValidatorStake:                 iotago.BaseToken(10),
			FixedCost:                      iotago.Mana(10),
			Active:                         true,
			LatestSupportedProtocolVersion: 1,
		},
		{
			AddressBech32:                  tpkg.RandAccountID().ToAddress().Bech32(iotago.PrefixTestnet),
			StakingEpochEnd:                iotago.EpochIndex(124),
			PoolStake:                      iotago.BaseToken(1000),
			ValidatorStake:                 iotago.BaseToken(100),
			FixedCost:                      iotago.Mana(20),
			Active:                         true,
			LatestSupportedProtocolVersion: 1,
		},
	}}

	mockGetJSON(api.CoreRouteValidators, 200, originRes)

	nodeAPI := nodeClient(t)
	res, err := nodeAPI.Validators(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_StakingByAccountID(t *testing.T) {
	defer gock.Off()

	accountAddress := tpkg.RandAccountID().ToAddress().(*iotago.AccountAddress)
	originRes := &api.ValidatorResponse{
		AddressBech32:                  accountAddress.Bech32(iotago.PrefixTestnet),
		StakingEpochEnd:                iotago.EpochIndex(123),
		PoolStake:                      iotago.BaseToken(100),
		ValidatorStake:                 iotago.BaseToken(10),
		FixedCost:                      iotago.Mana(10),
		Active:                         true,
		LatestSupportedProtocolVersion: 1,
	}

	nodeAPI := nodeClient(t)
	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteValidatorsAccount, api.ParameterBech32Address, accountAddress.Bech32(nodeAPI.CommittedAPI().ProtocolParameters().Bech32HRP())), 200, originRes)

	res, err := nodeAPI.StakingAccount(context.Background(), accountAddress)
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_Committee(t *testing.T) {
	defer gock.Off()

	originRes := &api.CommitteeResponse{
		Epoch:               iotago.EpochIndex(123),
		TotalStake:          1000_1000,
		TotalValidatorStake: 100_000,
		Committee: []*api.CommitteeMemberResponse{
			{
				AddressBech32:  tpkg.RandAccountID().ToAddress().Bech32(iotago.PrefixTestnet),
				PoolStake:      1000_000,
				ValidatorStake: 100_000,
				FixedCost:      iotago.Mana(100),
			},
		},
	}

	mockGetJSON(api.CoreRouteCommittee, 200, originRes)
	nodeAPI := nodeClient(t)
	res, err := nodeAPI.Committee(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, res)
}

func TestClient_SubmitBlock(t *testing.T) {
	defer gock.Off()

	blockHash := tpkg.Rand36ByteArray()
	blockHashStr := hexutil.EncodeHex(blockHash[:])

	incompleteBlock := &iotago.Block{
		API: mockAPI,
		Header: iotago.BlockHeader{
			ProtocolVersion:  mockAPI.Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(mockAPI).MustID(),
		},
		Signature: &iotago.Ed25519Signature{},
		Body: &iotago.BasicBlockBody{
			API:                mockAPI,
			StrongParents:      tpkg.SortedRandBlockIDs(1),
			WeakParents:        iotago.BlockIDs{},
			ShallowLikeParents: iotago.BlockIDs{},
		},
	}

	serializedIncompleteBlock, err := mockAPI.Encode(incompleteBlock, serix.WithValidation())
	require.NoError(t, err)

	gock.New(nodeAPIUrl).
		Post(api.CoreRouteBlocks).
		MatchType(api.MIMEApplicationVendorIOTASerializerV2).
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

	originRes := &api.BlockMetadataResponse{
		BlockID:          identifier,
		BlockState:       api.BlockStateConfirmed.String(),
		TransactionState: api.TransactionStateConfirmed.String(),
	}

	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteBlockMetadata, api.ParameterBlockID, identifier.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	meta, err := nodeAPI.BlockMetadataByBlockID(context.Background(), identifier)
	require.NoError(t, err)
	require.EqualValues(t, originRes, meta)
}

func TestClient_BlockByBlockID(t *testing.T) {
	defer gock.Off()

	blockID := tpkg.RandBlockID()

	originBlock := &iotago.Block{
		API: mockAPI,
		Header: iotago.BlockHeader{
			ProtocolVersion:  mockAPI.Version(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(mockAPI).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Body: &iotago.BasicBlockBody{
			API:                mockAPI,
			StrongParents:      tpkg.SortedRandBlockIDs(1 + rand.Intn(7)),
			WeakParents:        iotago.BlockIDs{},
			ShallowLikeParents: iotago.BlockIDs{},
			Payload:            nil,
		},
	}

	mockGetBinary(api.EndpointWithNamedParameterValue(api.CoreRouteBlock, api.ParameterBlockID, blockID.ToHex()), 200, originBlock)

	nodeAPI := nodeClient(t)
	responseBlock, err := nodeAPI.BlockByBlockID(context.Background(), blockID)
	require.NoError(t, err)
	require.EqualValues(t, lo.PanicOnErr(originBlock.ID()), lo.PanicOnErr(responseBlock.ID()))
}

func TestClient_TransactionIncludedBlock(t *testing.T) {
	defer gock.Off()

	txID := tpkg.RandTransactionID()

	originBlock := &iotago.Block{
		API: mockAPI,
		Header: iotago.BlockHeader{
			ProtocolVersion:  mockAPI.Version(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(mockAPI).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Body: &iotago.BasicBlockBody{
			API:                mockAPI,
			StrongParents:      tpkg.SortedRandBlockIDs(1 + rand.Intn(7)),
			WeakParents:        iotago.BlockIDs{},
			ShallowLikeParents: iotago.BlockIDs{},
			Payload:            nil,
		},
	}

	mockGetBinary(api.EndpointWithNamedParameterValue(api.CoreRouteTransactionsIncludedBlock, api.ParameterTransactionID, txID.ToHex()), 200, originBlock)

	nodeAPI := nodeClient(t)
	responseBlock, err := nodeAPI.TransactionIncludedBlock(context.Background(), txID)
	require.NoError(t, err)
	require.EqualValues(t, lo.PanicOnErr(originBlock.ID()), lo.PanicOnErr(responseBlock.ID()))
}

func TestClient_OutputByID(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)

	originOutputProof, err := iotago.NewOutputIDProof(tpkg.TestAPI, tpkg.Rand32ByteArray(), tpkg.RandSlot(), iotago.TxEssenceOutputs{originOutput}, 0)
	require.NoError(t, err)

	outputID, err := originOutputProof.OutputID(originOutput)
	require.NoError(t, err)

	mockGetBinary(api.EndpointWithNamedParameterValue(api.CoreRouteOutput, api.ParameterOutputID, outputID.ToHex()), 200, &api.OutputResponse{
		Output:        originOutput,
		OutputIDProof: originOutputProof,
	})

	nodeAPI := nodeClient(t)
	responseOutput, err := nodeAPI.OutputByID(context.Background(), outputID)
	require.NoError(t, err)

	require.EqualValues(t, originOutput, responseOutput)
}

func TestClient_OutputWithMetadataByID(t *testing.T) {
	defer gock.Off()

	originOutput := tpkg.RandBasicOutput(iotago.AddressEd25519)

	originOutputProof, err := iotago.NewOutputIDProof(tpkg.TestAPI, tpkg.Rand32ByteArray(), tpkg.RandSlot(), iotago.TxEssenceOutputs{originOutput}, 0)
	require.NoError(t, err)

	outputID, err := originOutputProof.OutputID(originOutput)
	require.NoError(t, err)

	originMetadata := &api.OutputMetadata{
		BlockID:              tpkg.RandBlockID(),
		TransactionID:        outputID.TransactionID(),
		OutputIndex:          outputID.Index(),
		IsSpent:              true,
		CommitmentIDSpent:    tpkg.Rand36ByteArray(),
		TransactionIDSpent:   tpkg.Rand36ByteArray(),
		IncludedCommitmentID: tpkg.Rand36ByteArray(),
		LatestCommitmentID:   tpkg.Rand36ByteArray(),
	}

	mockGetBinary(api.EndpointWithNamedParameterValue(api.CoreRouteOutputWithMetadata, api.ParameterOutputID, outputID.ToHex()), 200, &api.OutputWithMetadataResponse{
		Output:        originOutput,
		OutputIDProof: originOutputProof,
		Metadata:      originMetadata,
	})

	nodeAPI := nodeClient(t)
	responseOutput, responseMetadata, err := nodeAPI.OutputWithMetadataByID(context.Background(), outputID)
	require.NoError(t, err)

	require.EqualValues(t, originOutput, responseOutput)
	require.EqualValues(t, originMetadata, responseMetadata)
}

func TestClient_OutputMetadataByID(t *testing.T) {
	defer gock.Off()

	txID := tpkg.Rand36ByteArray()
	originRes := &api.OutputMetadata{
		BlockID:              tpkg.RandBlockID(),
		TransactionID:        txID,
		OutputIndex:          3,
		IsSpent:              true,
		CommitmentIDSpent:    tpkg.Rand36ByteArray(),
		TransactionIDSpent:   tpkg.Rand36ByteArray(),
		IncludedCommitmentID: tpkg.Rand36ByteArray(),
		LatestCommitmentID:   tpkg.Rand36ByteArray(),
	}

	utxoInput := &iotago.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
	utxoInputID := utxoInput.OutputID()

	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteOutputMetadata, api.ParameterOutputID, utxoInputID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.OutputMetadataByID(context.Background(), utxoInputID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)

	require.EqualValues(t, txID, resp.TransactionID)
}

func TestClient_CommitmentByID(t *testing.T) {
	defer gock.Off()

	var slot iotago.SlotIndex = 5

	commitmentID := iotago.NewCommitmentID(slot, tpkg.Rand32ByteArray())
	commitment := iotago.NewCommitment(mockAPI.Version(), slot, iotago.NewCommitmentID(slot-1, tpkg.Rand32ByteArray()), tpkg.Rand32ByteArray(), tpkg.RandUint64(math.MaxUint64), tpkg.RandMana(iotago.MaxMana))

	originRes := &iotago.Commitment{
		Slot:                 commitment.Slot,
		PreviousCommitmentID: commitment.PreviousCommitmentID,
		RootsID:              commitment.RootsID,
		CumulativeWeight:     commitment.CumulativeWeight,
	}

	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteCommitmentByID, api.ParameterCommitmentID, commitmentID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentByID(context.Background(), commitmentID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_CommitmentUTXOChangesByID(t *testing.T) {
	defer gock.Off()

	commitmentID := iotago.NewCommitmentID(5, tpkg.Rand32ByteArray())

	randCreatedOutput := tpkg.RandUTXOInput()
	randConsumedOutput := tpkg.RandUTXOInput()

	originRes := &api.UTXOChangesResponse{
		CommitmentID: commitmentID,
		CreatedOutputs: iotago.OutputIDs{
			randCreatedOutput.OutputID(),
		},
		ConsumedOutputs: iotago.OutputIDs{
			randConsumedOutput.OutputID(),
		},
	}

	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteCommitmentByIDUTXOChanges, api.ParameterCommitmentID, commitmentID.ToHex()), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentUTXOChangesByID(context.Background(), commitmentID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_CommitmentByIndex(t *testing.T) {
	defer gock.Off()

	var slot iotago.SlotIndex = 1337

	commitment := iotago.NewCommitment(mockAPI.Version(), slot, iotago.NewCommitmentID(slot-1, tpkg.Rand32ByteArray()), tpkg.Rand32ByteArray(), tpkg.RandUint64(math.MaxUint64), tpkg.RandMana(iotago.MaxMana))

	originRes := &iotago.Commitment{
		Slot:                 commitment.Slot,
		PreviousCommitmentID: commitment.PreviousCommitmentID,
		RootsID:              commitment.RootsID,
		CumulativeWeight:     commitment.CumulativeWeight,
	}

	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteCommitmentBySlot, api.ParameterSlot, strconv.Itoa(int(slot))), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentByIndex(context.Background(), slot)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestClient_CommitmentUTXOChangesByIndex(t *testing.T) {
	defer gock.Off()

	var slot iotago.SlotIndex = 1337
	commitmentID := iotago.NewCommitmentID(slot, tpkg.Rand32ByteArray())

	randCreatedOutput := tpkg.RandUTXOInput()
	randConsumedOutput := tpkg.RandUTXOInput()

	originRes := &api.UTXOChangesResponse{
		CommitmentID: commitmentID,
		CreatedOutputs: iotago.OutputIDs{
			randCreatedOutput.OutputID(),
		},
		ConsumedOutputs: iotago.OutputIDs{
			randConsumedOutput.OutputID(),
		},
	}

	mockGetJSON(api.EndpointWithNamedParameterValue(api.CoreRouteCommitmentBySlotUTXOChanges, api.ParameterSlot, strconv.Itoa(int(slot))), 200, originRes)

	nodeAPI := nodeClient(t)
	resp, err := nodeAPI.CommitmentUTXOChangesByIndex(context.Background(), slot)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

var sampleGossipInfo = &api.GossipInfo{
	Heartbeat: &api.GossipHeartbeat{
		SolidSlot:      234,
		PrunedSlot:     5872,
		LatestSlot:     1294,
		ConnectedPeers: 2392,
		SyncedPeers:    1234,
	},
	Metrics: &api.PeerGossipMetrics{
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
