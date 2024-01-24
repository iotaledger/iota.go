package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func TestTaggedDataDeSerialize(t *testing.T) {
	const tag = "寿司を作って"

	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok",
			Source: tpkg.RandTaggedData([]byte(tag)),
			Target: &iotago.TaggedData{},
		},
		{
			Name:   "empty-tag",
			Source: tpkg.RandTaggedData([]byte{}),
			Target: &iotago.TaggedData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
