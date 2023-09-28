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
		StorageCost:                             500,
		StorageScoreFactorData:                  1,
		StorageScoreOffsetOutput:                10,
		StorageScoreOffsetEd25519BlockIssuerKey: 50,
		StorageScoreOffsetStakingFeature:        100,
		StorageScoreOffsetDelegation:            100,
	}
	rentParametersJSON := `{"storageCost":"500","storageScoreFactorData":1,"storageScoreOffsetOutput":"10","storageScoreOffsetEd25519BlockIssuerKey":"50","storageScoreOffsetStakingFeature":"100","storageScoreOffsetDelegation":"100"}`

	j, err := tpkg.TestAPI.JSONEncode(rentParameters)
	require.NoError(t, err)

	require.Equal(t, rentParametersJSON, string(j))

	decodedRentStructure := &iotago.RentParameters{}
	err = tpkg.TestAPI.JSONDecode([]byte(rentParametersJSON), decodedRentStructure)
	require.NoError(t, err)

	require.Equal(t, rentParameters, decodedRentStructure)
}
