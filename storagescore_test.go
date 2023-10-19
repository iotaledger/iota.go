package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestStorageScoreParameters_DeSerialize(t *testing.T) {
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

func TestStorageScoreParamtersJSONMarshalling(t *testing.T) {
	storageScoreParameters := &iotago.StorageScoreParameters{
		StorageCost:                 500,
		FactorData:                  1,
		OffsetOutputOverhead:        10,
		OffsetEd25519BlockIssuerKey: 50,
		OffsetStakingFeature:        100,
		OffsetDelegation:            100,
	}
	storageScoreParametersJSON := `{"storageCost":"500","factorData":1,"offsetOutputOverhead":"10","offsetEd25519BlockIssuerKey":"50","offsetStakingFeature":"100","offsetDelegation":"100"}`

	j, err := tpkg.TestAPI.JSONEncode(storageScoreParameters)
	require.NoError(t, err)

	require.Equal(t, storageScoreParametersJSON, string(j))

	decodedStorageScoreStructure := &iotago.StorageScoreParameters{}
	err = tpkg.TestAPI.JSONDecode([]byte(storageScoreParametersJSON), decodedStorageScoreStructure)
	require.NoError(t, err)

	require.Equal(t, storageScoreParameters, decodedStorageScoreStructure)
}
