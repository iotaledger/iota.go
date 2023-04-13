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
	OneMi = 1_000_000
)

var (
	v3API = iotago.V3API(tpkg.TestProtoParams)
)

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
		TokenSupply:           1234567890987654321,
		GenesisUnixTimestamp:  1681373293,
		SlotDurationInSeconds: 10,
	}
	protoParamsJSON := `{"version":6,"networkName":"xxxNetwork","bech32Hrp":"xxx","minPowScore":666,"rentStructure":{"vByteCost":6,"vByteFactorData":8,"vByteFactorKey":7},"tokenSupply":"1234567890987654321","genesisUnixTimestamp":1681373293,"slotDurationInSeconds":10}`

	jsonProtoParams, err := v3API.JSONEncode(protoParams)
	require.NoError(t, err)
	require.Equal(t, protoParamsJSON, string(jsonProtoParams))

	decodedProtoParams := &iotago.ProtocolParameters{}
	err = v3API.JSONDecode([]byte(protoParamsJSON), decodedProtoParams)
	require.NoError(t, err)

	require.Equal(t, protoParams, decodedProtoParams)
}
