package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestRentParameters_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandRentParameters(),
			target: &iotago.RentParameters{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestRentParamtersJSONMarshalling(t *testing.T) {
	rentParameters := &iotago.RentParameters{
		VByteCost:                        500,
		VByteFactorData:                  1,
		VByteOffsetOutput:                10,
		VByteOffsetEd25519BlockIssuerKey: 50,
		VByteOffsetStakingFeature:        100,
		VByteOffsetDelegation:            100,
	}
	rentParametersJSON := `{"vByteCost":500,"vByteFactorData":1,"vByteOffsetOutput":"10","vByteOffsetEd25519BlockIssuerKey":"50","vByteOffsetStakingFeature":"100","vByteOffsetDelegation":"100"}`

	j, err := tpkg.TestAPI.JSONEncode(rentParameters)
	require.NoError(t, err)

	require.Equal(t, rentParametersJSON, string(j))

	decodedRentStructure := &iotago.RentParameters{}
	err = tpkg.TestAPI.JSONDecode([]byte(rentParametersJSON), decodedRentStructure)
	require.NoError(t, err)

	require.Equal(t, rentParameters, decodedRentStructure)
}
