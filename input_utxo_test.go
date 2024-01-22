package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestUTXOInput_DeSerialize(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name:   "",
			source: tpkg.RandUTXOInput(),
			target: &iotago.UTXOInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestUTXOInput_Equals(t *testing.T) {
	input1 := &iotago.UTXOInput{iotago.TransactionID{1, 2, 3, 4, 5, 6, 7}, 10}
	input2 := &iotago.UTXOInput{iotago.TransactionID{1, 2, 3, 4, 5, 6, 7}, 10}
	input3 := &iotago.UTXOInput{iotago.TransactionID{1, 2, 3, 4, 5, 6, 8}, 10}
	input4 := &iotago.UTXOInput{iotago.TransactionID{1, 2, 3, 4, 5, 6, 7}, 12}
	//nolint:gocritic // false positive
	require.True(t, input1.Equals(input1))
	require.True(t, input1.Equals(input2))
	require.False(t, input1.Equals(input3))
	require.False(t, input1.Equals(input4))
	require.False(t, input3.Equals(input4))
}
