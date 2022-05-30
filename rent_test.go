package iotago_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
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
		VByteCost:    500,
		VBFactorData: 1,
		VBFactorKey:  10,
	}
	rentStructureJSON := `{"vByteCost":500,"vByteFactorData":1,"vByteFactorKey":10}`

	j, err := json.Marshal(rentStructure)
	require.NoError(t, err)

	require.Equal(t, rentStructureJSON, string(j))

	decodedRentStructure := &iotago.RentStructure{}
	err = json.Unmarshal([]byte(rentStructureJSON), decodedRentStructure)
	require.NoError(t, err)

	require.Equal(t, rentStructure, decodedRentStructure)

	// Test encoding/decoding when used in other structs

	structBytes, err := json.Marshal(&struct {
		RentStructure iotago.RentStructure `json:"rentStructure"`
	}{
		RentStructure: *rentStructure,
	})
	require.NoError(t, err)

	expectedJSON := `{"rentStructure":{"vByteCost":500,"vByteFactorData":1,"vByteFactorKey":10}}`
	require.Equal(t, expectedJSON, string(structBytes))

	decodedStruct := &struct {
		RentStructure iotago.RentStructure `json:"rentStructure"`
	}{}

	err = json.Unmarshal(structBytes, decodedStruct)
	require.NoError(t, err)

	require.Equal(t, *rentStructure, decodedStruct.RentStructure)
}
