package api_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func Test_CoreAPIDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - InfoResponse",
			Source: &api.InfoResponse{
				Name:    "test",
				Version: "2.0.0",
				Status: &api.InfoResNodeStatus{
					IsHealthy:                   false,
					AcceptedTangleTime:          time.Unix(1690879505, 0).UTC(),
					RelativeAcceptedTangleTime:  time.Unix(1690879505, 0).UTC(),
					ConfirmedTangleTime:         time.Unix(1690879505, 0).UTC(),
					RelativeConfirmedTangleTime: time.Unix(1690879505, 0).UTC(),
					LatestCommitmentID:          tpkg.RandCommitmentID(),
					LatestFinalizedSlot:         tpkg.RandSlot(),
					LatestAcceptedBlockSlot:     tpkg.RandSlot(),
					LatestConfirmedBlockSlot:    tpkg.RandSlot(),
					PruningEpoch:                tpkg.RandEpoch(),
				},
				Metrics: &api.InfoResNodeMetrics{
					BlocksPerSecond:          1.1,
					ConfirmedBlocksPerSecond: 2.2,
					ConfirmationRate:         3.3,
				},
				ProtocolParameters: []*api.InfoResProtocolParameters{
					{
						StartEpoch: tpkg.RandEpoch(),
						Parameters: tpkg.RandProtocolParameters(),
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
			Target:    &api.InfoResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - IssuanceBlockHeaderResponse",
			Source: &api.IssuanceBlockHeaderResponse{
				StrongParents:                tpkg.SortedRandBlockIDs(2),
				WeakParents:                  tpkg.SortedRandBlockIDs(2),
				ShallowLikeParents:           tpkg.SortedRandBlockIDs(2),
				LatestParentBlockIssuingTime: time.Unix(1690879505, 0).UTC(),
				LatestFinalizedSlot:          tpkg.RandSlot(),
				LatestCommitment:             tpkg.RandCommitment(),
			},
			Target:    &api.IssuanceBlockHeaderResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - BlockCreatedResponse",
			Source: &api.BlockCreatedResponse{
				BlockID: tpkg.RandBlockID(),
			},
			Target:    &api.BlockCreatedResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - BlockMetadataResponse",
			Source: &api.BlockMetadataResponse{
				BlockID:    tpkg.RandBlockID(),
				BlockState: api.BlockStateDropped,
			},
			Target:    &api.BlockMetadataResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - BlockWithMetadataResponse",
			Source: &api.BlockWithMetadataResponse{
				Block: tpkg.RandBlock(tpkg.RandBasicBlockBody(tpkg.ZeroCostTestAPI, iotago.PayloadSignedTransaction), tpkg.ZeroCostTestAPI, 100),
				Metadata: &api.BlockMetadataResponse{
					BlockID:    tpkg.RandBlockID(),
					BlockState: api.BlockStateDropped,
				},
			},
			Target:    &api.BlockWithMetadataResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - TransactionMetadataResponse",
			Source: &api.TransactionMetadataResponse{
				TransactionID:             tpkg.RandTransactionID(),
				TransactionState:          api.TransactionStateFailed,
				EarliestAttachmentSlot:    5,
				TransactionFailureReason:  api.TxFailureDelegationRewardsClaimingInvalid,
				TransactionFailureDetails: "details",
			},
			Target:    &api.TransactionMetadataResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - OutputMetadata",
			Source: &api.OutputMetadata{
				OutputID: tpkg.RandOutputID(),
				BlockID:  tpkg.RandBlockID(),
				Included: &api.OutputInclusionMetadata{
					Slot:          tpkg.RandSlot(),
					TransactionID: tpkg.RandTransactionID(),
					CommitmentID:  tpkg.RandCommitmentID(),
				},
				Spent: &api.OutputConsumptionMetadata{
					Slot:          tpkg.RandSlot(),
					TransactionID: tpkg.RandTransactionID(),
					CommitmentID:  tpkg.RandCommitmentID(),
				},
				LatestCommitmentID: tpkg.RandCommitmentID(),
			},
			Target:    &api.OutputMetadata{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - OutputResponse",
			Source: &api.OutputResponse{
				Output:        tpkg.RandOutput(tpkg.RandOutputType()),
				OutputIDProof: tpkg.RandOutputIDProof(tpkg.ZeroCostTestAPI),
			},
			Target:    &api.OutputResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - OutputWithMetadataResponse",
			Source: &api.OutputWithMetadataResponse{
				Output:        tpkg.RandOutput(tpkg.RandOutputType()),
				OutputIDProof: tpkg.RandOutputIDProof(tpkg.ZeroCostTestAPI),
				Metadata: &api.OutputMetadata{
					OutputID: tpkg.RandOutputID(),
					BlockID:  tpkg.RandBlockID(),
					Included: &api.OutputInclusionMetadata{
						Slot:          tpkg.RandSlot(),
						TransactionID: tpkg.RandTransactionID(),
						CommitmentID:  tpkg.RandCommitmentID(),
					},
					Spent: &api.OutputConsumptionMetadata{
						Slot:          tpkg.RandSlot(),
						TransactionID: tpkg.RandTransactionID(),
						CommitmentID:  tpkg.RandCommitmentID(),
					},
					LatestCommitmentID: tpkg.RandCommitmentID(),
				},
			},
			Target:    &api.OutputWithMetadataResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - UTXOChangesResponse",
			Source: &api.UTXOChangesResponse{
				CommitmentID:    tpkg.RandCommitmentID(),
				CreatedOutputs:  tpkg.RandOutputIDs(3),
				ConsumedOutputs: tpkg.RandOutputIDs(3),
			},
			Target:    &api.UTXOChangesResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - UTXOChangesFullResponse",
			Source: &api.UTXOChangesFullResponse{
				CommitmentID: tpkg.RandCommitmentID(),
				CreatedOutputs: []*api.OutputWithID{
					{
						OutputID: tpkg.RandOutputID(),
						Output:   tpkg.RandOutput(tpkg.RandOutputType()),
					},
				},
				ConsumedOutputs: []*api.OutputWithID{
					{
						OutputID: tpkg.RandOutputID(),
						Output:   tpkg.RandOutput(tpkg.RandOutputType()),
					},
				},
			},
			Target:    &api.UTXOChangesFullResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - CongestionResponse",
			Source: &api.CongestionResponse{
				Slot:                 tpkg.RandSlot(),
				Ready:                true,
				ReferenceManaCost:    tpkg.RandMana(math.MaxUint32),
				BlockIssuanceCredits: 80,
			},
			Target:    &api.CongestionResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - ValidatorsResponse",
			Source: &api.ValidatorsResponse{
				Validators: []*api.ValidatorResponse{
					{
						AddressBech32:                  tpkg.RandAccountAddress().Bech32(iotago.PrefixTestnet),
						StakingEndEpoch:                tpkg.RandEpoch(),
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
			Target:    &api.ValidatorsResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - ManaRewardsResponse",
			Source: &api.ManaRewardsResponse{
				StartEpoch:                      tpkg.RandEpoch(),
				EndEpoch:                        tpkg.RandEpoch(),
				Rewards:                         456,
				LatestCommittedEpochPoolRewards: 555,
			},
			Target:    &api.ManaRewardsResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - CommitteeResponse",
			Source: &api.CommitteeResponse{
				Committee: []*api.CommitteeMemberResponse{
					{
						AddressBech32:  tpkg.RandAccountAddress().Bech32(iotago.PrefixTestnet),
						PoolStake:      456,
						ValidatorStake: 123,
						FixedCost:      789,
					},
				},
				TotalStake:          456,
				TotalValidatorStake: 123,
				Epoch:               tpkg.RandEpoch(),
			},
			Target:    &api.CommitteeResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
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

	tests := []*frameworks.JSONEncodeTest{
		{
			Name: "ok - InfoResponse",
			Source: &api.InfoResponse{
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
			Target: `{
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
					"dataByte": 500,
					"block": 110000,
					"input": 7500,
					"contextInput": 40000,
					"output": 90000,
					"nativeToken": 50000,
					"staking": 40000,
					"blockIssuer": 70000,
					"allotment": 5000,
					"signatureEd25519": 15000
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
					"increase": "1",
					"decrease": "1",
					"increaseThreshold": 400000000,
					"decreaseThreshold": 250000000,
					"schedulerRate": 50000000,
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
					"rewardToGenerationRatio": 2,
					"initialTargetRewardsRate": "616067521149261",
					"finalTargetRewardsRate": "226702563632670",
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
			Name: "ok - InfoResBaseToken - omitempty",
			Source: &api.InfoResBaseToken{
				Name:         "IOTA",
				TickerSymbol: "IOTA",
				Unit:         "MIOTA",
				// No Subunit
			},
			Target: `{
	"name": "IOTA",
	"tickerSymbol": "IOTA",
	"unit": "MIOTA",
	"decimals": 0
}`,
		},
		{
			Name: "ok - IssuanceBlockHeaderResponse",
			Source: &api.IssuanceBlockHeaderResponse{
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
			Target: `{
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
			Name: "ok - BlockCreatedResponse",
			Source: &api.BlockCreatedResponse{
				BlockID: iotago.BlockID{0x1},
			},
			Target: `{
	"blockId": "0x010000000000000000000000000000000000000000000000000000000000000000000000"
}`,
		},
		{
			Name: "ok - BlockMetadataResponse",
			Source: &api.BlockMetadataResponse{
				BlockID:    iotago.BlockID{0x9},
				BlockState: api.BlockStateDropped,
			},
			Target: `{
	"blockId": "0x090000000000000000000000000000000000000000000000000000000000000000000000",
	"blockState": "dropped"
}`,
		},
		{
			Name: "ok - BlockMetadataResponse - omitempty",
			Source: &api.BlockMetadataResponse{
				BlockID:    iotago.BlockID{0x9},
				BlockState: api.BlockStateConfirmed,
			},
			Target: `{
	"blockId": "0x090000000000000000000000000000000000000000000000000000000000000000000000",
	"blockState": "confirmed"
}`,
		},
		{
			Name: "ok - TransactionMetadataResponse",
			Source: &api.TransactionMetadataResponse{
				TransactionID:             iotago.TransactionID{0x1},
				TransactionState:          api.TransactionStateFailed,
				EarliestAttachmentSlot:    5,
				TransactionFailureReason:  api.TxFailureDelegationRewardsClaimingInvalid,
				TransactionFailureDetails: "details",
			},
			Target: `{
	"transactionId": "0x010000000000000000000000000000000000000000000000000000000000000000000000",
	"transactionState": "failed",
	"earliestAttachmentSlot": 5,
	"transactionFailureReason": 57,
	"transactionFailureDetails": "details"
}`,
		},
		{
			Name: "ok - TransactionMetadataResponse - omitempty",
			Source: &api.TransactionMetadataResponse{
				TransactionID:          iotago.TransactionID{0x1},
				TransactionState:       api.TransactionStateCommitted,
				EarliestAttachmentSlot: 10,
			},
			Target: `{
	"transactionId": "0x010000000000000000000000000000000000000000000000000000000000000000000000",
	"transactionState": "committed",
	"earliestAttachmentSlot": 10
}`,
		},
		{
			Name: "ok - OutputMetadata",
			Source: &api.OutputMetadata{
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
			Target: `{
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
			Name: "ok - OutputMetadata - omitempty",
			Source: &api.OutputMetadata{
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
			Target: `{
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
			Name: "ok - UTXOChangesResponse",
			Source: &api.UTXOChangesResponse{
				CommitmentID: iotago.NewCommitmentID(42, iotago.Identifier{}),
				CreatedOutputs: iotago.OutputIDs{
					iotago.OutputID{0x1},
				},
				ConsumedOutputs: iotago.OutputIDs{
					iotago.OutputID{0x2},
				},
			},
			Target: `{
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
			Name: "ok - UTXOChangesFullResponse",
			Source: &api.UTXOChangesFullResponse{
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
			Target: `{
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
			Name: "ok - CongestionResponse",
			Source: &api.CongestionResponse{
				Slot:                 12,
				Ready:                true,
				ReferenceManaCost:    100,
				BlockIssuanceCredits: 80,
			},
			Target: `{
	"slot": 12,
	"ready": true,
	"referenceManaCost": "100",
	"blockIssuanceCredits": "80"
}`,
		},
		{
			Name: "ok - ValidatorsResponse",
			Source: &api.ValidatorsResponse{
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
			Target: `{
	"validators": [
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
			Name: "ok - ManaRewardsResponse",
			Source: &api.ManaRewardsResponse{
				StartEpoch:                      123,
				EndEpoch:                        133,
				Rewards:                         456,
				LatestCommittedEpochPoolRewards: 555,
			},
			Target: `{
	"startEpoch": 123,
	"endEpoch": 133,
	"rewards": "456",
	"latestCommittedEpochPoolRewards": "555"
}`,
		},
		{
			Name: "ok - CommitteeResponse",
			Source: &api.CommitteeResponse{
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
			Target: `{
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
		t.Run(tt.Name, tt.Run)
	}
}

func TestTransactionFailureReasonDetermination(t *testing.T) {
	type txFailureTest struct {
		name     string
		err      error
		expected api.TransactionFailureReason
	}

	tests := []*txFailureTest{
		{
			name:     "last error of a series of joined errors is mapped first",
			err:      ierrors.Join(iotago.ErrAccountLocked, iotago.ErrDelegationAmountMismatch, iotago.ErrBlockIssuerCommitmentInputMissing),
			expected: api.TxFailureBlockIssuerCommitmentInputMissing,
		},
		{
			name: "first visited error of a post-order depth traversed error tree is mapped first",
			err: func() error {
				err1 := ierrors.WithMessage(iotago.ErrAccountLocked, "message1")
				errTree1 := ierrors.WithMessage(err1, "message2")

				subtreeErr := ierrors.WithMessage(iotago.ErrRewardInputReferenceInvalid, "message3")
				subTree := ierrors.WithMessage(subtreeErr, "message4")

				err2 := ierrors.WithMessage(iotago.ErrAccountInvalidFoundryCounter, "message5")
				errTree2 := ierrors.Join(err2, subTree)

				return ierrors.Join(errTree1, errTree2)
			}(),
			expected: api.TxFailureRewardInputReferenceInvalid,
		},
	}

	for _, test := range tests {
		err := test.err
		expected := test.expected

		t.Run(test.name, func(t *testing.T) {
			txFailureReason := api.DetermineTransactionFailureReason(err)
			require.Equal(t, expected, txFailureReason,
				"expected %d, got %d", expected, txFailureReason)
		})
	}
}
