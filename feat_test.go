package iotago_test

import (
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestFeaturesDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - StakingFeature",
			source: &iotago.StakingFeature{
				StakedAmount: 100,
				FixedCost:    12,
				StartEpoch:   100,
				EndEpoch:     1236,
			},
			target: &iotago.StakingFeature{},
		},
		{
			name: "ok - BlockIssuerFeature",
			source: &iotago.BlockIssuerFeature{
				BlockIssuerKeys: iotago.BlockIssuerKeys{
					iotago.BlockIssuerKeyEd25519FromPublicKey(ed25519.PublicKey(tpkg.RandBytes(32))),
				},
				ExpirySlot: 10,
			},
			target: &iotago.BlockIssuerFeature{},
		},
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
