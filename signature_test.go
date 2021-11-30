package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
)

func TestEd25519Signature_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandEd25519Signature(),
			target: &iotago.Ed25519Signature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
