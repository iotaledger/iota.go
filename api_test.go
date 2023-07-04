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

var v3API = iotago.V3API(tpkg.TestProtoParams)

type deSerializeTest struct {
	name      string
	source    any
	target    any
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	serixData, err := v3API.Encode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.Error(t, err, test.seriErr)
		return
	}
	require.NoError(t, err)

	if src, ok := test.source.(iotago.Sizer); ok {
		require.Equal(t, src.Size(), len(serixData))
	}

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := v3API.Decode(serixData, serixTarget)
	if test.deSeriErr != nil {
		require.Error(t, err, test.deSeriErr)
		return
	}
	require.NoError(t, err)
	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	sourceJson, err := v3API.JSONEncode(test.source)
	require.NoError(t, err)

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, v3API.JSONDecode(sourceJson, jsonDest))

	require.EqualValues(t, test.source, jsonDest)
}

func TestProtocolParameters_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandProtocolParameters(),
			target: &iotago.ProtocolParameters{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestProtocolParametersJSONMarshalling(t *testing.T) {
	protoParams := &iotago.ProtocolParameters{
		Version:     6,
		NetworkName: "xxxNetwork",
		Bech32HRP:   "xxx",
		MinPoWScore: 666,
		RentStructure: iotago.RentStructure{
			VByteCost:    6,
			VBFactorKey:  7,
			VBFactorData: 8,
		},
		TokenSupply:                1234567890987654321,
		GenesisUnixTimestamp:       1681373293,
		SlotDurationInSeconds:      10,
		SlotsPerEpochExponent:      13,
		ManaGenerationRate:         1,
		ManaGenerationRateExponent: 27,
		ManaDecayFactors: []uint32{
			10,
			20,
		},
		ManaDecayFactorsExponent:         32,
		ManaDecayFactorEpochsSum:         1337,
		ManaDecayFactorEpochsSumExponent: 20,
		EvictionAge:                      10,
		StakingUnbondingPeriod:           11,
		LivenessThreshold:                3,
		WorkScoreStructure: iotago.WorkScoreStructure{
			FactorData:                1,
			FactorInput:               2,
			FactorAllotment:           3,
			FactorMissingParent:       4,
			WorkScoreOutput:           5,
			WorkScoreStaking:          6,
			WorkScoreBlockIssuer:      7,
			WorkScoreEd25519Signature: 8,
			WorkScoreNativeToken:      9,
			WorkScoreMaxParents:       10,
		},
		EpochNearingThreshold: 4,
	}
	protoParamsJSON := `{"version":6,"networkName":"xxxNetwork","bech32Hrp":"xxx","minPowScore":666,"rentStructure":{"vByteCost":6,"vByteFactorData":8,"vByteFactorKey":7},"tokenSupply":"1234567890987654321","genesisUnixTimestamp":"1681373293","slotDurationInSeconds":10,"slotsPerEpochExponent":13,"manaGenerationRate":1,"manaGenerationRateExponent":27,"manaDecayFactors":[10,20],"manaDecayFactorsExponent":32,"manaDecayFactorEpochsSum":1337,"manaDecayFactorEpochsSumExponent":20,"stakingUnbondingPeriod":"11","evictionAge":"10","liveNessThreshold":"3","workScoreStructure":{"factorData":1,"factorInput":2,"factorAllotment":3,"factorMissingParent":4,"workScoreOutput":"5","workScoreStaking":"6","workScoreBlockIssuer":"7","workScoreEd25519Signature":"8","workScoreNativeToken":"9","workScoreMaxParents":"10"},"epochNearingThreshold":"0","epochNearingThreshold":"4"}`

	jsonProtoParams, err := v3API.JSONEncode(protoParams)
	require.NoError(t, err)
	require.Equal(t, protoParamsJSON, string(jsonProtoParams))

	decodedProtoParams := &iotago.ProtocolParameters{}
	err = v3API.JSONDecode([]byte(protoParamsJSON), decodedProtoParams)
	require.NoError(t, err)

	require.Equal(t, protoParams, decodedProtoParams)
}
