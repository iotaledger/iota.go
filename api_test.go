//nolint:scopelint
package iotago_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	OneIOTA iotago.BaseToken = 1_000_000
)

type deSerializeTest struct {
	name      string
	source    any
	target    any
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) assertBinaryEncodeDecode(t *testing.T) {
	t.Helper()

	serixData, err := tpkg.ZeroCostTestAPI.Encode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "binary encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		serixData, err = tpkg.ZeroCostTestAPI.Encode(test.source)
		require.NoError(t, err, "binary encoding")
	} else {
		require.NoError(t, err, "binary encoding")
	}

	if src, ok := test.source.(iotago.Sizer); ok {
		require.Equal(t, src.Size(), len(serixData), "binary encoding")
	}

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr, "binary decoding")

		return
	}
	require.NoError(t, err, "binary decoding")
	require.Len(t, serixData, bytesRead, "binary decoding")
	require.EqualValues(t, test.source, serixTarget, "binary decoding")
}

func (test *deSerializeTest) assertJSONEncodeDecode(t *testing.T) {
	t.Helper()

	sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "JSON encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		sourceJSON, err = tpkg.ZeroCostTestAPI.JSONEncode(test.source)
		require.NoError(t, err, "JSON encoding")
	} else {
		require.NoError(t, err, "JSON encoding")
	}

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	err = tpkg.ZeroCostTestAPI.JSONDecode(sourceJSON, jsonDest, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr, "JSON decoding")

		return
	}
	require.NoError(t, err, "JSON decoding")
	require.EqualValues(t, test.source, jsonDest, "JSON decoding")
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	t.Helper()

	if reflect.TypeOf(test.target).Kind() != reflect.Ptr {
		// This is required for the serixTarget creation hack to work.
		t.Fatal("test target must be a pointer")
	}

	test.assertBinaryEncodeDecode(t)
	test.assertJSONEncodeDecode(t)
}

func TestProtocolParameters_DeSerialize(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name:      "ok",
			source:    tpkg.RandProtocolParameters(),
			target:    &iotago.V3ProtocolParameters{},
			seriErr:   nil,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

// jsonEncodeTest is used to check if the JSON encoding is equal to a manually provided JSON string.
type jsonEncodeTest struct {
	name   string
	source any
	// the target should be an indented JSON string (tabs instead of spaces)
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

	tests := []*jsonEncodeTest{
		{
			name:   "ok",
			source: protoParams,
			target: `{
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
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
