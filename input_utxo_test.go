package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
)

func TestUTXOInput_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
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
