//nolint:scopelint
package iotago_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	OneMi iotago.BaseToken = 1_000_000
)

type deSerializeTest struct {
	name      string
	source    any
	target    any
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	serixData, err := tpkg.TestAPI.Encode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr)

		return
	}
	require.NoError(t, err)

	if src, ok := test.source.(iotago.Sizer); ok {
		require.Equal(t, src.Size(), len(serixData))
	}

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := tpkg.TestAPI.Decode(serixData, serixTarget)
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr)

		return
	}
	require.NoError(t, err)
	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	sourceJSON, err := tpkg.TestAPI.JSONEncode(test.source)
	require.NoError(t, err)

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, tpkg.TestAPI.JSONDecode(sourceJSON, jsonDest))

	require.EqualValues(t, test.source, jsonDest)
}

func TestProtocolParameters_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandProtocolParameters(),
			target: &iotago.V3ProtocolParameters{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestProtocolParametersJSONMarshalling(t *testing.T) {
	var schedulerRate iotago.WorkScore = 100000
	var protoParams iotago.ProtocolParameters = iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions(
			"xxxNetwork",
			"xxx",
		),
		iotago.WithSupplyOptions(
			1234567890987654321,
			6,
			7,
			8,
			9,
			10,
			10,
		),
		iotago.WithWorkScoreOptions(
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
		),
		iotago.WithTimeProviderOptions(
			1681373293,
			10,
			13,
		),
		iotago.WithManaOptions(
			1,
			1,
			27,
			[]uint32{10, 20},
			32,
			1337,
			20,
		),
		iotago.WithStakingOptions(11, 10, 9),
		iotago.WithLivenessOptions(
			15,
			30,
			10,
			20,
			24,
		),
		iotago.WithCongestionControlOptions(
			500,
			500,
			500,
			8*schedulerRate, // 0.8*slotDurationInSeconds*schedulerRate
			5*schedulerRate, // 0.5*slotDurationInSeconds*schedulerRate
			schedulerRate,
			1000,
			100,
		),
		iotago.WithVersionSignalingOptions(3, 4, 1),
		iotago.WithRewardsOptions(8, 8, 31, 1154, 2, 1),
	)

	protoParamsJSON := `{"type":0,"version":3,"networkName":"xxxNetwork","bech32Hrp":"xxx","rentParameters":{"storageCost":"6","storageScoreFactorData":7,"storageScoreOffsetOutput":"8","storageScoreOffsetEd25519BlockIssuerKey":"9","storageScoreOffsetStakingFeature":"10","storageScoreOffsetDelegation":"10"},"workScoreParameters":{"dataByte":1,"block":2,"input":3,"contextInput":4,"output":5,"nativeToken":6,"staking":7,"blockIssuer":8,"allotment":9,"signatureEd25519":10},"tokenSupply":"1234567890987654321","genesisUnixTimestamp":"1681373293","slotDurationInSeconds":10,"slotsPerEpochExponent":13,"manaStructure":{"bitsCount":1,"generationRate":1,"generationRateExponent":27,"decayFactors":[10,20],"decayFactorsExponent":32,"decayFactorEpochsSum":1337,"decayFactorEpochsSumExponent":20},"stakingUnbondingPeriod":11,"validationBlocksPerSlot":10,"punishmentEpochs":9,"livenessThresholdLowerBound":15,"livenessThresholdUpperBound":30,"minCommittableAge":10,"maxCommittableAge":20,"epochNearingThreshold":24,"congestionControlParameters":{"minReferenceManaCost":"500","increase":"500","decrease":"500","increaseThreshold":800000,"decreaseThreshold":500000,"schedulerRate":100000,"maxBufferSize":1000,"maxValidationBufferSize":100},"versionSignaling":{"windowSize":3,"windowTargetRatio":4,"activationOffset":1},"rewardsParameters":{"profitMarginExponent":8,"bootstrappingDuration":1154,"manaShareCoefficient":"2","decayBalancingConstantExponent":8,"decayBalancingConstant":"1","poolCoefficientExponent":31}}`

	jsonProtoParams, err := tpkg.TestAPI.JSONEncode(protoParams)
	require.NoError(t, err)
	require.Equal(t, protoParamsJSON, string(jsonProtoParams))

	var decodedProtoParams iotago.ProtocolParameters
	err = tpkg.TestAPI.JSONDecode([]byte(protoParamsJSON), &decodedProtoParams)
	require.NoError(t, err)

	require.Equal(t, protoParams, decodedProtoParams)
}
