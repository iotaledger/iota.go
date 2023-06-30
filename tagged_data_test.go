package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTaggedDataDeSerialize(t *testing.T) {
	const tag = "寿司を作って"

	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTaggedData([]byte(tag)),
			target: &iotago.TaggedData{},
		},
		{
			name:   "empty-tag",
			source: tpkg.RandTaggedData(nil),
			target: &iotago.TaggedData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
