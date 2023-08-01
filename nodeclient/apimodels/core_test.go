package apimodels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

func testAPI() iotago.API {
	params := iotago.NewV3ProtocolParameters(
		iotago.WithTimeProviderOptions(time.Unix(1690879505, 0).Unix(), 10, 13),
	)

	api := iotago.V3API(params)
	return api
}

func Test_InfoResponse(t *testing.T) {
	api := testAPI()
	{
		response := &apimodels.InfoResponse{
			Name:    "test",
			Version: "2.0.0",
			Status: &apimodels.InfoResNodeStatus{
				IsHealthy:                   false,
				AcceptedTangleTime:          time.Unix(1690879505, 0),
				RelativeAcceptedTangleTime:  time.Unix(1690879505, 0),
				ConfirmedTangleTime:         time.Unix(1690879505, 0),
				RelativeConfirmedTangleTime: time.Unix(1690879505, 0),
				LatestCommitmentID:          iotago.CommitmentID{},
				LatestFinalizedSlot:         1,
				LatestAcceptedBlockSlot:     2,
				LatestConfirmedBlockSlot:    3,
				PruningSlot:                 4,
			},
			Metrics: &apimodels.InfoResNodeMetrics{
				BlocksPerSecond:          1.1,
				ConfirmedBlocksPerSecond: 2.2,
				ConfirmationRate:         3.3,
			},
			ProtocolParameters: []*apimodels.InfoResProtocolParameters{
				{
					StartEpoch: 0,
					Parameters: api.ProtocolParameters(),
				},
			},
			BaseToken: &apimodels.InfoResBaseToken{
				Name:            "Shimmer",
				TickerSymbol:    "SMR",
				Unit:            "SMR",
				Subunit:         "glow",
				Decimals:        6,
				UseMetricPrefix: false,
			},
			Features: []string{"test"},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"name\":\"test\",\"version\":\"2.0.0\",\"status\":{\"isHealthy\":false,\"acceptedTangleTime\":\"2023-08-01T10:45:05+02:00\",\"relativeAcceptedTangleTime\":\"2023-08-01T10:45:05+02:00\",\"confirmedTangleTime\":\"2023-08-01T10:45:05+02:00\",\"relativeConfirmedTangleTime\":\"2023-08-01T10:45:05+02:00\",\"latestCommitmentId\":\"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"latestFinalizedSlot\":\"1\",\"latestAcceptedBlockSlot\":\"2\",\"latestConfirmedBlockSlot\":\"3\",\"pruningSlot\":\"4\"},\"metrics\":{\"blocksPerSecond\":\"1.1E+00\",\"confirmedBlocksPerSecond\":\"2.2E+00\",\"confirmationRate\":\"3.3E+00\"},\"protocolParameters\":[{\"startEpoch\":\"0\",\"parameters\":{\"type\":0,\"version\":3,\"networkName\":\"testnet\",\"bech32Hrp\":\"rms\",\"rentStructure\":{\"vByteCost\":100,\"vByteFactorData\":1,\"vByteFactorKey\":10},\"workScoreStructure\":{\"dataByte\":1,\"block\":100,\"missingParent\":500,\"input\":20,\"contextInput\":20,\"output\":20,\"nativeToken\":20,\"staking\":100,\"blockIssuer\":100,\"allotment\":100,\"signatureEd25519\":200,\"minStrongParentsThreshold\":4},\"tokenSupply\":\"1813620509061365\",\"genesisUnixTimestamp\":\"1690879505\",\"slotDurationInSeconds\":10,\"slotsPerEpochExponent\":13,\"manaGenerationRate\":1,\"manaGenerationRateExponent\":0,\"manaDecayFactors\":[10,20],\"manaDecayFactorsExponent\":0,\"manaDecayFactorEpochsSum\":0,\"manaDecayFactorEpochsSumExponent\":0,\"stakingUnbondingPeriod\":\"10\",\"evictionAge\":\"10\",\"livenessThreshold\":\"3\",\"epochNearingThreshold\":\"4\",\"versionSignaling\":{\"windowSize\":7,\"windowTargetRatio\":5,\"activationOffset\":7}}}],\"baseToken\":{\"name\":\"Shimmer\",\"tickerSymbol\":\"SMR\",\"unit\":\"SMR\",\"subunit\":\"glow\",\"decimals\":6,\"useMetricPrefix\":false},\"features\":[\"test\"]}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.InfoResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &apimodels.InfoResBaseToken{
			Name:            "IOTA",
			TickerSymbol:    "IOTA",
			Unit:            "MIOTA",
			UseMetricPrefix: true,
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"name\":\"IOTA\",\"tickerSymbol\":\"IOTA\",\"unit\":\"MIOTA\",\"decimals\":0,\"useMetricPrefix\":true}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.InfoResBaseToken)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_IssuanceBlockHeaderResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.IssuanceBlockHeaderResponse{
		StrongParents: iotago.BlockIDs{
			iotago.BlockID{0x9},
		},
		WeakParents: iotago.BlockIDs{
			iotago.BlockID{0x8},
		},
		ShallowLikeParents: iotago.BlockIDs{
			iotago.BlockID{0x7},
		},
		LatestFinalizedSlot: 14,
		Commitment: &iotago.Commitment{
			Version:          api.Version(),
			Index:            18,
			PrevID:           iotago.CommitmentID{0x1},
			RootsID:          iotago.Identifier{0x2},
			CumulativeWeight: 89,
		},
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"strongParents\":[\"0x09000000000000000000000000000000000000000000000000000000000000000000000000000000\"],\"weakParents\":[\"0x08000000000000000000000000000000000000000000000000000000000000000000000000000000\"],\"shallowLikeParents\":[\"0x07000000000000000000000000000000000000000000000000000000000000000000000000000000\"],\"latestFinalizedSlot\":\"14\",\"commitment\":{\"version\":3,\"index\":\"18\",\"prevId\":\"0x01000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"rootsId\":\"0x0200000000000000000000000000000000000000000000000000000000000000\",\"cumulativeWeight\":\"89\"}}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.IssuanceBlockHeaderResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_BlockCreatedResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.BlockCreatedResponse{
		BlockID: iotago.BlockID{0x1},
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"blockId\":\"0x01000000000000000000000000000000000000000000000000000000000000000000000000000000\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.BlockCreatedResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_BlockMetadataResponse(t *testing.T) {
	api := testAPI()

	{
		response := &apimodels.BlockMetadataResponse{
			BlockID:            iotago.BlockID{0x9},
			BlockState:         apimodels.BlockStateFailed.String(),
			BlockFailureReason: apimodels.BlockFailureBookingFailure,
			TxState:            apimodels.TransactionStateFailed.String(),
			TxFailureReason:    apimodels.TxFailureFailedToClaimDelegationReward,
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x09000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"blockState\":\"failed\",\"blockFailureReason\":3,\"txState\":\"failed\",\"txFailureReason\":21}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.BlockMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &apimodels.BlockMetadataResponse{
			BlockID:    iotago.BlockID{0x9},
			BlockState: apimodels.BlockStateConfirmed.String(),
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x09000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"blockState\":\"confirmed\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.BlockMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_OutputMetadataResponse(t *testing.T) {
	api := testAPI()

	{
		response := &apimodels.OutputMetadataResponse{
			BlockID:              iotago.BlockID{0x8},
			TransactionID:        iotago.TransactionID{0x9},
			OutputIndex:          3,
			IsSpent:              true,
			CommitmentIDSpent:    iotago.CommitmentID{0x6},
			TransactionIDSpent:   iotago.TransactionID{0x1},
			IncludedCommitmentID: iotago.CommitmentID{0x3},
			LatestCommitmentID:   iotago.CommitmentID{0x2},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x08000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"transactionId\":\"0x0900000000000000000000000000000000000000000000000000000000000000\",\"outputIndex\":3,\"isSpent\":true,\"commitmentIdSpent\":\"0x06000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"transactionIdSpent\":\"0x0100000000000000000000000000000000000000000000000000000000000000\",\"includedCommitmentId\":\"0x03000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"latestCommitmentId\":\"0x02000000000000000000000000000000000000000000000000000000000000000000000000000000\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.OutputMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &apimodels.OutputMetadataResponse{
			BlockID:            iotago.BlockID{0x8},
			TransactionID:      iotago.TransactionID{0x9},
			OutputIndex:        3,
			IsSpent:            false,
			LatestCommitmentID: iotago.CommitmentID{0x2},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x08000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"transactionId\":\"0x0900000000000000000000000000000000000000000000000000000000000000\",\"outputIndex\":3,\"isSpent\":false,\"latestCommitmentId\":\"0x02000000000000000000000000000000000000000000000000000000000000000000000000000000\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.OutputMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_UTXOChangesResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.UTXOChangesResponse{
		Index: 42,
		CreatedOutputs: iotago.OutputIDs{
			iotago.OutputID{0x1},
		},
		ConsumedOutputs: iotago.OutputIDs{
			iotago.OutputID{0x2},
		},
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"index\":\"42\",\"createdOutputs\":[\"0x01000000000000000000000000000000000000000000000000000000000000000000\"],\"consumedOutputs\":[\"0x02000000000000000000000000000000000000000000000000000000000000000000\"]}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.UTXOChangesResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_CongestionResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.CongestionResponse{
		SlotIndex:            12,
		Ready:                true,
		ReferenceManaCost:    100,
		BlockIssuanceCredits: 80,
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"slotIndex\":\"12\",\"ready\":true,\"referenceManaCost\":\"100\",\"blockIssuanceCredits\":\"80\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.CongestionResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_AccountStakingListResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.AccountStakingListResponse{
		Stakers: []*apimodels.ValidatorResponse{
			{
				AccountID:                      iotago.AccountID{0xFF},
				StakingEpochEnd:                0,
				PoolStake:                      123,
				ValidatorStake:                 456,
				FixedCost:                      69,
				LatestSupportedProtocolVersion: 9,
			},
		},
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"stakers\":[{\"accountId\":\"0xff00000000000000000000000000000000000000000000000000000000000000\",\"stakingEpochEnd\":\"0\",\"poolStake\":\"123\",\"validatorStake\":\"456\",\"fixedCost\":\"69\",\"latestSupportedProtocolVersion\":9}]}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.AccountStakingListResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_ManaRewardsResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.ManaRewardsResponse{
		EpochIndex: 123,
		Rewards:    456,
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"epochIndex\":\"123\",\"rewards\":\"456\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.ManaRewardsResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_CommitteeResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.CommitteeResponse{
		Committee: []*apimodels.CommitteeMemberResponse{
			{
				AccountID:      iotago.AccountID{0xFF},
				PoolStake:      456,
				ValidatorStake: 123,
				FixedCost:      789,
			},
		},
		TotalStake:          456,
		TotalValidatorStake: 123,
		EpochIndex:          872,
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"committee\":[{\"accountId\":\"0xff00000000000000000000000000000000000000000000000000000000000000\",\"poolStake\":\"456\",\"validatorStake\":\"123\",\"fixedCost\":\"789\"}],\"totalStake\":\"456\",\"totalValidatorStake\":\"123\",\"epochIndex\":\"872\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.CommitteeResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}