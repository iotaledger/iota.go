package api_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func testAPI() iotago.API {
	params := tpkg.FixedGenesisV3TestProtocolParameters

	return iotago.V3API(params)
}

// jsonEncodeTest is used to check if the JSON encoding is equal to a manually provided JSON string.
type jsonEncodeTest struct {
	name   string
	source any
	// the target should be an indented JSON string (4 white spaces)
	target string
}

func (test *jsonEncodeTest) run(t *testing.T) {
	t.Helper()

	sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(test.source, serix.WithValidation())
	require.NoError(t, err, "JSON encoding")

	var b bytes.Buffer
	err = json.Indent(&b, sourceJSON, "", "\t")
	require.NoError(t, err, "JSON indenting")
	indentedJSON := b.String()

	require.EqualValues(t, test.target, string(indentedJSON))
}

func Test_CoreAPIJSONSerialization(t *testing.T) {

	protoParams := iotago.NewV3SnapshotProtocolParameters(
		iotago.WithTimeProviderOptions(1, 1690879505, 10, 13),
	)

	// replace the decay factors to reduce the size of the JSON string
	protoParams.ManaParameters().DecayFactors = []uint32{
		1,
		2,
		3,
	}

	tests := []*jsonEncodeTest{
		{
			name: "ok - InfoResponse",
			source: &api.InfoResponse{
				Name:    "test",
				Version: "2.0.0",
				Status: &api.InfoResNodeStatus{
					IsHealthy:                   false,
					AcceptedTangleTime:          time.Unix(1690879505, 0).UTC(),
					RelativeAcceptedTangleTime:  time.Unix(1690879505, 0).UTC(),
					ConfirmedTangleTime:         time.Unix(1690879505, 0).UTC(),
					RelativeConfirmedTangleTime: time.Unix(1690879505, 0).UTC(),
					LatestCommitmentID:          iotago.CommitmentID{},
					LatestFinalizedSlot:         1,
					LatestAcceptedBlockSlot:     2,
					LatestConfirmedBlockSlot:    3,
					PruningEpoch:                4,
				},
				Metrics: &api.InfoResNodeMetrics{
					BlocksPerSecond:          1.1,
					ConfirmedBlocksPerSecond: 2.2,
					ConfirmationRate:         3.3,
				},
				ProtocolParameters: []*api.InfoResProtocolParameters{
					{
						StartEpoch: 0,
						Parameters: protoParams,
					},
				},
				BaseToken: &api.InfoResBaseToken{
					Name:         "Shimmer",
					TickerSymbol: "SMR",
					Unit:         "SMR",
					Subunit:      "glow",
					Decimals:     6,
				},
			},
			target: `{
	"name": "test",
	"version": "2.0.0",
	"status": {
		"isHealthy": false,
		"acceptedTangleTime": "1690879505000000000",
		"relativeAcceptedTangleTime": "1690879505000000000",
		"confirmedTangleTime": "1690879505000000000",
		"relativeConfirmedTangleTime": "1690879505000000000",
		"latestCommitmentId": "0x000000000000000000000000000000000000000000000000000000000000000000000000",
		"latestFinalizedSlot": 1,
		"latestAcceptedBlockSlot": 2,
		"latestConfirmedBlockSlot": 3,
		"pruningEpoch": 4
	},
	"metrics": {
		"blocksPerSecond": "1.1E+00",
		"confirmedBlocksPerSecond": "2.2E+00",
		"confirmationRate": "3.3E+00"
	},
	"protocolParameters": [
		{
			"startEpoch": 0,
			"parameters": {
				"type": 0,
				"version": 3,
				"networkName": "testnet",
				"bech32Hrp": "rms",
				"storageScoreParameters": {
					"storageCost": "100",
					"factorData": 1,
					"offsetOutputOverhead": "10",
					"offsetEd25519BlockIssuerKey": "100",
					"offsetStakingFeature": "100",
					"offsetDelegation": "100"
				},
				"workScoreParameters": {
					"dataByte": 1,
					"block": 100,
					"input": 10,
					"contextInput": 20,
					"output": 20,
					"nativeToken": 20,
					"staking": 5000,
					"blockIssuer": 1000,
					"allotment": 1000,
					"signatureEd25519": 1000
				},
				"manaParameters": {
					"bitsCount": 63,
					"generationRate": 1,
					"generationRateExponent": 17,
					"decayFactors": [
						1,
						2,
						3
					],
					"decayFactorsExponent": 32,
					"decayFactorEpochsSum": 2262417561,
					"decayFactorEpochsSumExponent": 21,
					"annualDecayFactorPercentage": 70
				},
				"tokenSupply": "1813620509061365",
				"genesisSlot": 1,
				"genesisUnixTimestamp": "1690879505",
				"slotDurationInSeconds": 10,
				"slotsPerEpochExponent": 13,
				"stakingUnbondingPeriod": 10,
				"validationBlocksPerSlot": 10,
				"punishmentEpochs": 10,
				"livenessThresholdLowerBound": 15,
				"livenessThresholdUpperBound": 30,
				"minCommittableAge": 10,
				"maxCommittableAge": 20,
				"epochNearingThreshold": 60,
				"congestionControlParameters": {
					"minReferenceManaCost": "1",
					"increase": "10",
					"decrease": "10",
					"increaseThreshold": 800000,
					"decreaseThreshold": 500000,
					"schedulerRate": 100000,
					"maxBufferSize": 1000,
					"maxValidationBufferSize": 100
				},
				"versionSignalingParameters": {
					"windowSize": 7,
					"windowTargetRatio": 5,
					"activationOffset": 7
				},
				"rewardsParameters": {
					"profitMarginExponent": 8,
					"bootstrappingDuration": 1079,
					"manaShareCoefficient": "2",
					"decayBalancingConstantExponent": 8,
					"decayBalancingConstant": "1",
					"poolCoefficientExponent": 11,
					"retentionPeriod": 384
				},
				"targetCommitteeSize": 32,
				"chainSwitchingThreshold": 3
			}
		}
	],
	"baseToken": {
		"name": "Shimmer",
		"tickerSymbol": "SMR",
		"unit": "SMR",
		"subunit": "glow",
		"decimals": 6
	}
}`,
		},
		{
			name: "ok - InfoResBaseToken - omitempty",
			source: &api.InfoResBaseToken{
				Name:         "IOTA",
				TickerSymbol: "IOTA",
				Unit:         "MIOTA",
				// No Subunit
			},
			target: `{
	"name": "IOTA",
	"tickerSymbol": "IOTA",
	"unit": "MIOTA",
	"decimals": 0
}`,
		},
		{
			name: "ok - IssuanceBlockHeaderResponse",
			source: &api.IssuanceBlockHeaderResponse{
				StrongParents: iotago.BlockIDs{
					iotago.BlockID{0x9},
				},
				WeakParents: iotago.BlockIDs{
					iotago.BlockID{0x8},
				},
				ShallowLikeParents: iotago.BlockIDs{
					iotago.BlockID{0x7},
				},
				LatestParentBlockIssuingTime: time.Unix(1690879505, 0).UTC(),
				LatestFinalizedSlot:          14,
				LatestCommitment: &iotago.Commitment{
					ProtocolVersion:      3,
					Slot:                 18,
					PreviousCommitmentID: iotago.CommitmentID{0x1},
					RootsID:              iotago.Identifier{0x2},
					CumulativeWeight:     89,
					ReferenceManaCost:    123,
				},
			},
			target: `{
	"strongParents": [
		"0x090000000000000000000000000000000000000000000000000000000000000000000000"
	],
	"weakParents": [
		"0x080000000000000000000000000000000000000000000000000000000000000000000000"
	],
	"shallowLikeParents": [
		"0x070000000000000000000000000000000000000000000000000000000000000000000000"
	],
	"latestParentBlockIssuingTime": "1690879505000000000",
	"latestFinalizedSlot": 14,
	"latestCommitment": {
		"protocolVersion": 3,
		"slot": 18,
		"previousCommitmentId": "0x010000000000000000000000000000000000000000000000000000000000000000000000",
		"rootsId": "0x0200000000000000000000000000000000000000000000000000000000000000",
		"cumulativeWeight": "89",
		"referenceManaCost": "123"
	}
}`,
		},
		{
			name: "ok - BlockCreatedResponse",
			source: &api.BlockCreatedResponse{
				BlockID: iotago.BlockID{0x1},
			},
			target: `{
	"blockId": "0x010000000000000000000000000000000000000000000000000000000000000000000000"
}`,
		},
		{
			name: "ok - BlockMetadataResponse",
			source: &api.BlockMetadataResponse{
				BlockID:            iotago.BlockID{0x9},
				BlockState:         api.BlockStateFailed,
				BlockFailureReason: api.BlockFailureParentNotFound,
				TransactionMetadata: &api.TransactionMetadataResponse{
					TransactionID:            iotago.TransactionID{0x1},
					TransactionState:         api.TransactionStateFailed,
					TransactionFailureReason: api.TxFailureFailedToClaimDelegationReward,
				},
			},
			target: `{
	"blockId": "0x090000000000000000000000000000000000000000000000000000000000000000000000",
	"blockState": "failed",
	"blockFailureReason": 3,
	"transactionMetadata": {
		"transactionId": "0x010000000000000000000000000000000000000000000000000000000000000000000000",
		"transactionState": "failed",
		"transactionFailureReason": 20
	}
}`,
		},
		{
			name: "ok - BlockMetadataResponse - omitempty",
			source: &api.BlockMetadataResponse{
				BlockID:    iotago.BlockID{0x9},
				BlockState: api.BlockStateConfirmed,
			},
			target: `{
	"blockId": "0x090000000000000000000000000000000000000000000000000000000000000000000000",
	"blockState": "confirmed"
}`,
		},
		{
			name: "ok - OutputMetadata",
			source: &api.OutputMetadata{
				OutputID: iotago.OutputID{0x01},
				BlockID:  iotago.BlockID{0x02},
				Included: &api.OutputInclusionMetadata{
					Slot:          3,
					TransactionID: iotago.TransactionID{0x4},
					CommitmentID:  iotago.CommitmentID{0x5},
				},
				Spent: &api.OutputConsumptionMetadata{
					Slot:          6,
					TransactionID: iotago.TransactionID{0x7},
					CommitmentID:  iotago.CommitmentID{0x8},
				},
				LatestCommitmentID: iotago.CommitmentID{0x9},
			},
			target: `{
	"outputId": "0x0100000000000000000000000000000000000000000000000000000000000000000000000000",
	"blockId": "0x020000000000000000000000000000000000000000000000000000000000000000000000",
	"included": {
		"slot": 3,
		"transactionId": "0x040000000000000000000000000000000000000000000000000000000000000000000000",
		"commitmentId": "0x050000000000000000000000000000000000000000000000000000000000000000000000"
	},
	"spent": {
		"slot": 6,
		"transactionId": "0x070000000000000000000000000000000000000000000000000000000000000000000000",
		"commitmentId": "0x080000000000000000000000000000000000000000000000000000000000000000000000"
	},
	"latestCommitmentId": "0x090000000000000000000000000000000000000000000000000000000000000000000000"
}`,
		},
		{
			name: "ok - OutputMetadata - omitempty",
			source: &api.OutputMetadata{
				OutputID: iotago.OutputID{0x01},
				BlockID:  iotago.BlockID{0x02},
				Included: &api.OutputInclusionMetadata{
					Slot:          3,
					TransactionID: iotago.TransactionID{0x4},
					// CommitmentID is omitted
				},
				// Spent is omitted
				LatestCommitmentID: iotago.CommitmentID{0x9},
			},
			target: `{
	"outputId": "0x0100000000000000000000000000000000000000000000000000000000000000000000000000",
	"blockId": "0x020000000000000000000000000000000000000000000000000000000000000000000000",
	"included": {
		"slot": 3,
		"transactionId": "0x040000000000000000000000000000000000000000000000000000000000000000000000"
	},
	"latestCommitmentId": "0x090000000000000000000000000000000000000000000000000000000000000000000000"
}`,
		},
		{
			name: "ok - UTXOChangesResponse",
			source: &api.UTXOChangesResponse{
				CommitmentID: iotago.NewCommitmentID(42, iotago.Identifier{}),
				CreatedOutputs: iotago.OutputIDs{
					iotago.OutputID{0x1},
				},
				ConsumedOutputs: iotago.OutputIDs{
					iotago.OutputID{0x2},
				},
			},
			target: `{
	"commitmentId": "0x00000000000000000000000000000000000000000000000000000000000000002a000000",
	"createdOutputs": [
		"0x0100000000000000000000000000000000000000000000000000000000000000000000000000"
	],
	"consumedOutputs": [
		"0x0200000000000000000000000000000000000000000000000000000000000000000000000000"
	]
}`,
		},
		{
			name: "ok - UTXOChangesFullResponse",
			source: &api.UTXOChangesFullResponse{
				CommitmentID: iotago.NewCommitmentID(42, iotago.Identifier{}),
				CreatedOutputs: []*api.OutputWithID{
					{
						OutputID: iotago.OutputID{0x1},
						Output: &iotago.BasicOutput{
							Amount: 123,
							Mana:   456,
							UnlockConditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{
									Address: &iotago.Ed25519Address{0x01},
								},
							},
							Features: iotago.BasicOutputFeatures{},
						},
					},
				},
				ConsumedOutputs: []*api.OutputWithID{
					{
						OutputID: iotago.OutputID{0x2},
						Output: &iotago.BasicOutput{
							Amount: 456,
							Mana:   123,
							UnlockConditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{
									Address: &iotago.Ed25519Address{0x02},
								},
							},
							Features: iotago.BasicOutputFeatures{},
						},
					},
				},
			},
			target: `{
	"commitmentId": "0x00000000000000000000000000000000000000000000000000000000000000002a000000",
	"createdOutputs": [
		{
			"outputId": "0x0100000000000000000000000000000000000000000000000000000000000000000000000000",
			"output": {
				"type": 0,
				"amount": "123",
				"mana": "456",
				"unlockConditions": [
					{
						"type": 0,
						"address": {
							"type": 0,
							"pubKeyHash": "0x0100000000000000000000000000000000000000000000000000000000000000"
						}
					}
				]
			}
		}
	],
	"consumedOutputs": [
		{
			"outputId": "0x0200000000000000000000000000000000000000000000000000000000000000000000000000",
			"output": {
				"type": 0,
				"amount": "456",
				"mana": "123",
				"unlockConditions": [
					{
						"type": 0,
						"address": {
							"type": 0,
							"pubKeyHash": "0x0200000000000000000000000000000000000000000000000000000000000000"
						}
					}
				]
			}
		}
	]
}`,
		},
		{
			name: "ok - CongestionResponse",
			source: &api.CongestionResponse{
				Slot:                 12,
				Ready:                true,
				ReferenceManaCost:    100,
				BlockIssuanceCredits: 80,
			},
			target: `{
	"slot": 12,
	"ready": true,
	"referenceManaCost": "100",
	"blockIssuanceCredits": "80"
}`,
		},
		{
			name: "ok - ValidatorsResponse",
			source: &api.ValidatorsResponse{
				Validators: []*api.ValidatorResponse{
					{
						AddressBech32:                  iotago.AccountID{0xFF}.ToAddress().Bech32(iotago.PrefixTestnet),
						StakingEndEpoch:                0,
						PoolStake:                      123,
						ValidatorStake:                 456,
						FixedCost:                      69,
						Active:                         true,
						LatestSupportedProtocolVersion: 9,
					},
				},
				Cursor:   "0,1",
				PageSize: 50,
			},
			target: `{
	"stakers": [
		{
			"address": "rms1prlsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqcyz9fx",
			"stakingEndEpoch": 0,
			"poolStake": "123",
			"validatorStake": "456",
			"fixedCost": "69",
			"active": true,
			"latestSupportedProtocolVersion": 9,
			"latestSupportedProtocolHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
		}
	],
	"pageSize": 50,
	"cursor": "0,1"
}`,
		},
		{
			name: "ok - ManaRewardsResponse",
			source: &api.ManaRewardsResponse{
				StartEpoch:                      123,
				EndEpoch:                        133,
				Rewards:                         456,
				LatestCommittedEpochPoolRewards: 555,
			},
			target: `{
	"startEpoch": 123,
	"endEpoch": 133,
	"rewards": "456",
	"latestCommittedEpochPoolRewards": "555"
}`,
		},
		{
			name: "ok - CommitteeResponse",
			source: &api.CommitteeResponse{
				Committee: []*api.CommitteeMemberResponse{
					{
						AddressBech32:  iotago.AccountID{0xFF}.ToAddress().Bech32(iotago.PrefixTestnet),
						PoolStake:      456,
						ValidatorStake: 123,
						FixedCost:      789,
					},
				},
				TotalStake:          456,
				TotalValidatorStake: 123,
				Epoch:               872,
			},
			target: `{
	"committee": [
		{
			"address": "rms1prlsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqcyz9fx",
			"poolStake": "456",
			"validatorStake": "123",
			"fixedCost": "789"
		}
	],
	"totalStake": "456",
	"totalValidatorStake": "123",
	"epoch": 872
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
