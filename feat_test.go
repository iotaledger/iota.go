package iotago_test

import (
	"reflect"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/stretchr/testify/require"
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
			seriErr:   iotago.ErrInvalidMetadataKey,
			deSeriErr: iotago.ErrInvalidMetadataKey,
			target:    &iotago.MetadataFeature{},
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
			seriErr:   iotago.ErrInvalidStateMetadataKey,
			deSeriErr: iotago.ErrInvalidStateMetadataKey,
			target:    &iotago.StateMetadataFeature{},
		},
		{
			name: "fail - StateMetadataFeature - space char in key",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"space-> ": []byte("world"),
				},
			},
			seriErr:   iotago.ErrInvalidStateMetadataKey,
			deSeriErr: iotago.ErrInvalidStateMetadataKey,
			target:    &iotago.StateMetadataFeature{},
		},
		{
			name: "fail - StateMetadataFeature - ASCII control-character in key",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"\x07": []byte("world"),
				},
			},
			seriErr:   iotago.ErrInvalidStateMetadataKey,
			deSeriErr: iotago.ErrInvalidStateMetadataKey,
			target:    &iotago.StateMetadataFeature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

// Tests that maps are sorted when encoded to binary to produce a deterministic result,
// but do not have to be sorted when decoded from binary
func TestFeaturesMetadataLexicalOrdering(t *testing.T) {
	type metadataDeserializeTest struct {
		name   string
		source iotago.Feature
		target iotago.Feature
	}

	tests := []metadataDeserializeTest{
		{
			name: "ok - MetadataFeature",
			source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"b": []byte("y"),
					"c": []byte("z"),
					"a": []byte("x"),
				},
			},
			target: &iotago.MetadataFeature{},
		},
		{
			name: "ok - StateMetadataFeature",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"b": []byte("y"),
					"c": []byte("z"),
					"a": []byte("x"),
				},
			},
			target: &iotago.StateMetadataFeature{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serixData, err := tpkg.ZeroCostTestAPI.Encode(test.source, serix.WithValidation())
			require.NoError(t, err)

			expected := []byte{
				// Metadata Feature Type
				byte(test.source.Type()),
				// Map Length
				3,
				// Key Length
				1,
				'a',
				// Little-endian value Length
				1, 0,
				'x',
				// Key Length
				1,
				'b',
				// Little-endian value Length
				1, 0,
				'y',
				// Key Length
				1,
				'c',
				// Little-endian value Length
				1, 0,
				'z',
			}

			require.Equal(t, expected, serixData)

			// Decoding the sorted map should succeed.
			bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, test.target, serix.WithValidation())
			require.NoError(t, err)
			require.Len(t, serixData, bytesRead)
			require.EqualValues(t, test.source, test.target)

			// Swap a and b to make it unsorted.
			serixData[3], serixData[8] = serixData[8], serixData[3]
			// Swap x and y so the maps are equal key-value-wise.
			serixData[6], serixData[11] = serixData[11], serixData[6]

			// Decoding the unsorted map should fail.
			serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
			_, err = tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())
			require.ErrorIs(t, err, serializer.ErrArrayValidationOrderViolatesLexicalOrder)
		})
	}
}
