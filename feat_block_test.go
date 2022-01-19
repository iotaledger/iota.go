package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestFeatureBlocksDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - SenderFeatureBlock",
			source: &iotago.SenderFeatureBlock{Address: tpkg.RandEd25519Address()},
			target: &iotago.SenderFeatureBlock{},
		},
		{
			name:   "ok - IssuerFeatureBlock",
			source: &iotago.IssuerFeatureBlock{Address: tpkg.RandEd25519Address()},
			target: &iotago.IssuerFeatureBlock{},
		},
		{
			name: "ok - MetadataFeatureBlock",
			source: &iotago.MetadataFeatureBlock{
				Data: []byte("hello world"),
			},
			target: &iotago.MetadataFeatureBlock{},
		},
		{
			name: "ok - TagFeatureBlock",
			source: &iotago.TagFeatureBlock{
				Tag: []byte("hello world"),
			},
			target: &iotago.TagFeatureBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
