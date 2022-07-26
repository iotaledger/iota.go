package iotago_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

const (
	OneMi = 1_000_000
)

type deSerializeTest struct {
	name      string
	source    serializer.Serializable
	target    serializer.Serializable
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	data, err := test.source.Serialize(serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	if test.seriErr != nil {
		require.Error(t, err, test.seriErr)
		return
	}
	assert.NoError(t, err)
	if src, ok := test.source.(serializer.SerializableWithSize); ok {
		assert.Equal(t, len(data), src.Size())
	}

	bytesRead, err := test.target.Deserialize(data, serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	if test.deSeriErr != nil {
		require.Error(t, err, test.deSeriErr)
		return
	}
	assert.NoError(t, err)
	require.Len(t, data, bytesRead)
	assert.EqualValues(t, test.source, test.target)
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
	protoParas := &iotago.ProtocolParameters{
		Version:       6,
		NetworkName:   "xxxNetwork",
		Bech32HRP:     "xxx",
		MinPoWScore:   666,
		BelowMaxDepth: 15,
		RentStructure: iotago.RentStructure{
			VByteCost:    6,
			VBFactorKey:  7,
			VBFactorData: 8,
		},
		TokenSupply: 1234567890987654321,
	}
	protoParasJSON := `{"version":6,"networkName":"xxxNetwork","bech32Hrp":"xxx","minPowScore":666,"belowMaxDepth":15,"rentStructure":{"vByteCost":6,"vByteFactorData":8,"vByteFactorKey":7},"tokenSupply":"1234567890987654321"}`

	j, err := json.Marshal(protoParas)
	require.NoError(t, err)

	require.Equal(t, protoParasJSON, string(j))

	decodedProtoParas := &iotago.ProtocolParameters{}
	err = json.Unmarshal([]byte(protoParasJSON), decodedProtoParas)
	require.NoError(t, err)

	require.Equal(t, protoParas, decodedProtoParas)

	// Test encoding/decoding when used in other structs

	structBytes, err := json.Marshal(&struct {
		ProtocolParameters iotago.ProtocolParameters `json:"protocol"`
	}{
		ProtocolParameters: *protoParas,
	})
	require.NoError(t, err)

	expectedJSON := `{"protocol":{"version":6,"networkName":"xxxNetwork","bech32Hrp":"xxx","minPowScore":666,"belowMaxDepth":15,"rentStructure":{"vByteCost":6,"vByteFactorData":8,"vByteFactorKey":7},"tokenSupply":"1234567890987654321"}}`
	require.Equal(t, expectedJSON, string(structBytes))

	decodedStruct := &struct {
		ProtocolParameters iotago.ProtocolParameters `json:"protocol"`
	}{}

	err = json.Unmarshal(structBytes, decodedStruct)
	require.NoError(t, err)

	require.Equal(t, *protoParas, decodedStruct.ProtocolParameters)
}
