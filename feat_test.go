package iotago_test

import (
	"testing"

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
				BlockIssuerKeys: iotago.NewBlockIssuerKeys(
					iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
				),
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
				Entries: iotago.MetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"":         []byte(""),
				},
			},
			target: &iotago.MetadataFeature{},
		},
		{
			name: "ok - StateMetadataFeature",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"":         []byte(""),
				},
			},
			target: &iotago.StateMetadataFeature{},
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

func TestFeaturesMetadata(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - MetadataFeature",
			source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"empty":    []byte(""),
				},
			},
			target: &iotago.MetadataFeature{},
		},
		{
			name: "fail - MetadataFeature - non ASCII char in key",
			source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hellö": []byte("world"),
				},
			},
			seriErr: iotago.ErrInvalidMetadataKey,
			target:  &iotago.MetadataFeature{},
		},
		{
			name: "ok - StateMetadataFeature",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"empty":    []byte(""),
				},
			},
			target: &iotago.StateMetadataFeature{},
		},
		{
			name: "fail - StateMetadataFeature - non ASCII char in key",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hellö": []byte("world"),
				},
			},
			seriErr: iotago.ErrInvalidStateMetadataKey,
			target:  &iotago.StateMetadataFeature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
