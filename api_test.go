//nolint:scopelint
package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

const (
	OneIOTA iotago.BaseToken = 1_000_000
)

func TestProtocolParameters_DeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:      "ok",
			Source:    tpkg.RandProtocolParameters(),
			Target:    &iotago.V3ProtocolParameters{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestProtocolParametersJSONMarshalling(t *testing.T) {

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
			Name:   "ok",
			Source: protoParams,
			Target: `{
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
		"increase": "10",
		"decrease": "10",
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
		"manaShareCoefficient": "2",
		"decayBalancingConstantExponent": 8,
		"decayBalancingConstant": "1",
		"poolCoefficientExponent": 11,
		"retentionPeriod": 384
	},
	"targetCommitteeSize": 32,
	"chainSwitchingThreshold": 3
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
