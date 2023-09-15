package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestRentStructure_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandRentStructure(),
			target: &iotago.RentStructure{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestRentStructureJSONMarshalling(t *testing.T) {
	rentStructure := &iotago.RentStructure{
		VByteCost:              500,
		VBFactorData:           1,
		VBFactorKey:            10,
		VBFactorIssuerKeys:     50,
		VBFactorStakingFeature: 100,
		VBFactorDelegation:     100,
	}
	rentStructureJSON := `{"vByteCost":500,"vByteFactorData":1,"vByteFactorKey":10,"vByteFactorIssuerKeys":50,"vByteFactorStakingFeature":100,"vByteFactorDelegation":100}`

	j, err := tpkg.TestAPI.JSONEncode(rentStructure)
	require.NoError(t, err)

	require.Equal(t, rentStructureJSON, string(j))

	decodedRentStructure := &iotago.RentStructure{}
	err = tpkg.TestAPI.JSONDecode([]byte(rentStructureJSON), decodedRentStructure)
	require.NoError(t, err)

	require.Equal(t, rentStructure, decodedRentStructure)
}
