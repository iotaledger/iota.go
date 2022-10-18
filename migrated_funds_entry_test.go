package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v4/tpkg"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestMigratedFundsEntry_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandMigratedFundsEntry(),
			target: &iotago.MigratedFundsEntry{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
