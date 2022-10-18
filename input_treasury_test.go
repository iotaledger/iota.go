package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v4/tpkg"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestTreasuryInput_Deserialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTreasuryInput(),
			target: &iotago.TreasuryInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
