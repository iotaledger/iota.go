package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
)

func TestIndexationDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandIndexation(),
			target: &iotago.Indexation{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
