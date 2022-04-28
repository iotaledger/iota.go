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

	mapRentStructure, err := v2API.MapEncode(rentStructure)
	require.NoError(t, err)
	j, err := json.Marshal(mapRentStructure)
	require.NoError(t, err)

	require.Equal(t, rentStructureJSON, string(j))

	decodedRentStructure := &iotago.RentStructure{}
	m := map[string]any{}
	err = json.Unmarshal([]byte(rentStructureJSON), &m)
	require.NoError(t, err)
	err = v2API.MapDecode(m, decodedRentStructure)
	require.NoError(t, err)

	require.Equal(t, rentStructure, decodedRentStructure)
}
