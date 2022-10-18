package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestFeaturesDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - SenderFeature",
			source: &iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
			target: &iotago.SenderFeature{},
		},
		{
			name:   "ok - Issuer",
			source: &iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
			target: &iotago.IssuerFeature{},
		},
		{
			name: "ok - MetadataFeature",
			source: &iotago.MetadataFeature{
				Data: []byte("hello world"),
			},
			target: &iotago.MetadataFeature{},
		},
		{
			name: "ok - TagFeature",
			source: &iotago.TagFeature{
				Tag: []byte("hello world"),
			},
			target: &iotago.TagFeature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
