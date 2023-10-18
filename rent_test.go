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
			source: tpkg.RandStorageScoreParameters(),
			target: &iotago.StorageScoreParameters{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestRentParamtersJSONMarshalling(t *testing.T) {
	rentParameters := &iotago.StorageScoreParameters{
		StorageCost:                 500,
		FactorData:                  1,
		OffsetOutputOverhead:        10,
		OffsetEd25519BlockIssuerKey: 50,
		OffsetStakingFeature:        100,
		OffsetDelegation:            100,
	}
	rentParametersJSON := `{"storageCost":"500","factorData":1,"offsetOutputOverhead":"10","offsetEd25519BlockIssuerKey":"50","offsetStakingFeature":"100","offsetDelegation":"100"}`

	j, err := tpkg.TestAPI.JSONEncode(rentParameters)
	require.NoError(t, err)

	require.Equal(t, rentParametersJSON, string(j))

	decodedRentStructure := &iotago.StorageScoreParameters{}
	err = tpkg.TestAPI.JSONDecode([]byte(rentParametersJSON), decodedRentStructure)
	require.NoError(t, err)

	require.Equal(t, rentParameters, decodedRentStructure)
}
