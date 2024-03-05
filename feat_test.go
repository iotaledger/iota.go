package iotago_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func TestFeaturesDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - StakingFeature",
			Source: &iotago.StakingFeature{
				StakedAmount: 100,
				FixedCost:    12,
				StartEpoch:   100,
				EndEpoch:     1236,
			},
			Target: &iotago.StakingFeature{},
		},
		{
			Name: "ok - BlockIssuerFeature",
			Source: &iotago.BlockIssuerFeature{
				BlockIssuerKeys: iotago.NewBlockIssuerKeys(
					iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
				),
				ExpirySlot: 10,
			},
			Target: &iotago.BlockIssuerFeature{},
		},
		{
			Name:   "ok - SenderFeature",
			Source: &iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
			Target: &iotago.SenderFeature{},
		},
		{
			Name:   "ok - Issuer",
			Source: &iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
			Target: &iotago.IssuerFeature{},
		},
		{
			Name: "ok - MetadataFeature",
			Source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"":         []byte(""),
				},
			},
			Target: &iotago.MetadataFeature{},
		},
		{
			Name: "ok - StateMetadataFeature",
			Source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"":         []byte(""),
				},
			},
			Target: &iotago.StateMetadataFeature{},
		},
		{
			Name: "ok - TagFeature",
			Source: &iotago.TagFeature{
				Tag: []byte("hello world"),
			},
			Target: &iotago.TagFeature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestFeaturesMetadata(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - MetadataFeature",
			Source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"empty":    []byte(""),
				},
			},
			Target: &iotago.MetadataFeature{},
		},
		{
			Name: "fail - MetadataFeature - non ASCII char in key",
			Source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hellö": []byte("world"),
				},
			},
			SeriErr:   iotago.ErrInvalidMetadataKey,
			DeSeriErr: iotago.ErrInvalidMetadataKey,
			Target:    &iotago.MetadataFeature{},
		},
		{
			Name: "ok - StateMetadataFeature",
			Source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"empty":    []byte(""),
				},
			},
			Target: &iotago.StateMetadataFeature{},
		},
		{
			Name: "fail - StateMetadataFeature - non ASCII char in key",
			Source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hellö": []byte("world"),
				},
			},
			SeriErr:   iotago.ErrInvalidStateMetadataKey,
			DeSeriErr: iotago.ErrInvalidStateMetadataKey,
			Target:    &iotago.StateMetadataFeature{},
		},
		{
			Name: "fail - StateMetadataFeature - space char in key",
			Source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"space-> ": []byte("world"),
				},
			},
			SeriErr:   iotago.ErrInvalidStateMetadataKey,
			DeSeriErr: iotago.ErrInvalidStateMetadataKey,
			Target:    &iotago.StateMetadataFeature{},
		},
		{
			Name: "fail - StateMetadataFeature - ASCII control-character in key",
			Source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"\x07": []byte("world"),
				},
			},
			SeriErr:   iotago.ErrInvalidStateMetadataKey,
			DeSeriErr: iotago.ErrInvalidStateMetadataKey,
			Target:    &iotago.StateMetadataFeature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

// Tests that maps are sorted when encoded to and decoded from binary to produce a deterministic result,
// but do not have to be sorted when encoded/decoded to JSON.
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
		// Required to avoid triggering the scopelint.
		source := test.source
		target := test.target
		featType := test.source.Type()

		t.Run(test.name, func(t *testing.T) {
			{
				serixData, err := tpkg.ZeroCostTestAPI.Encode(source, serix.WithValidation())
				require.NoError(t, err)

				expected := []byte{
					// Metadata Feature Type
					byte(featType),
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
				bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, target, serix.WithValidation())
				require.NoError(t, err)
				require.Len(t, serixData, bytesRead)
				require.EqualValues(t, source, target)

				// Swap a and b to make it unsorted.
				serixData[3], serixData[8] = serixData[8], serixData[3]
				// Swap x and y so the maps are equal key-value-wise.
				serixData[6], serixData[11] = serixData[11], serixData[6]

				// Decoding the unsorted map should fail.
				serixTarget := reflect.New(reflect.TypeOf(target).Elem()).Interface()
				_, err = tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())
				require.ErrorIs(t, err, serializer.ErrArrayValidationOrderViolatesLexicalOrder)
			}

			{
				sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(source, serix.WithValidation())
				require.NoError(t, err)

				json := string(sourceJSON)
				require.Contains(t, json, fmt.Sprintf(`"type":%d`, byte(source.Type())))
				require.Contains(t, json, `"a":"0x78"`)
				require.Contains(t, json, `"b":"0x79"`)
				require.Contains(t, json, `"c":"0x7a"`)

				sortedJSON := fmt.Sprintf(`{"type":%d,"entries":{"a":"0x78","b":"0x79","c":"0x7a"}}`, byte(source.Type()))
				unsortedJSON := fmt.Sprintf(`{"type":%d,"entries":{"b":"0x79","a":"0x78","c":"0x7a"}}`, byte(source.Type()))

				// Both sorted and unsorted input is accepted.
				for _, src := range []string{sortedJSON, unsortedJSON} {
					serixTarget := reflect.New(reflect.TypeOf(target).Elem()).Interface()
					err = tpkg.ZeroCostTestAPI.JSONDecode([]byte(src), serixTarget, serix.WithValidation())
					require.NoError(t, err)
					require.Equal(t, source, serixTarget)
				}
			}
		})
	}
}

func TestMetadataMaxSize(t *testing.T) {
	myKey := "mykey"
	myKeyLen := len(myKey)
	mapLenPrefixSize := 1
	keyLenPrefixSize := 1
	valueLenPrefixSize := 2

	tests := []transactionSerializeTest{
		{
			name: "ok - MetadataFeature size matches max allowed size",
			output: func() iotago.Output {
				output := &iotago.BasicOutput{
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
					},
				}
				output.Amount = 100_000_000
				output.Features = append(output.Features, &iotago.MetadataFeature{
					Entries: iotago.MetadataFeatureEntries{
						iotago.MetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
							iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize,
						),
					},
				})

				return output
			}(),
		},
		{
			name: "fail - MetadataFeature size exceeds max allowed size by one",
			output: func() iotago.Output {
				output := tpkg.RandBasicOutput()
				output.Amount = 100_000_000
				output.Features = append(output.Features, &iotago.MetadataFeature{
					Entries: iotago.MetadataFeatureEntries{
						iotago.MetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
							iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize + 1,
						),
					},
				})

				return output
			}(),
			seriErr:   iotago.ErrMetadataExceedsMaxSize,
			deseriErr: iotago.ErrMetadataExceedsMaxSize,
		},
		{
			name: "ok - StateMetadataFeature size matches max allowed size",
			output: func() iotago.Output {
				return &iotago.AnchorOutput{
					Amount: 100_000_000,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
						&iotago.GovernorAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
					},
					ImmutableFeatures: iotago.AnchorOutputImmFeatures{},
					Features: iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{
							Entries: iotago.StateMetadataFeatureEntries{
								iotago.StateMetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
									iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize,
								),
							},
						},
					},
				}
			}(),
		},
		{
			name: "fail - StateMetadataFeature size exceeds max allowed size by one",
			output: func() iotago.Output {
				return &iotago.AnchorOutput{
					Amount: 100_000_000,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
						&iotago.GovernorAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
					},
					ImmutableFeatures: iotago.AnchorOutputImmFeatures{},
					Features: iotago.AnchorOutputFeatures{
						&iotago.MetadataFeature{
							Entries: iotago.MetadataFeatureEntries{
								"test": []byte("value_unrelated_to_test"),
							},
						},
						&iotago.StateMetadataFeature{
							Entries: iotago.StateMetadataFeatureEntries{
								iotago.StateMetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
									iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize + 1,
								),
							},
						},
					},
				}
			}(),
			seriErr:   iotago.ErrMetadataExceedsMaxSize,
			deseriErr: iotago.ErrMetadataExceedsMaxSize,
		},
	}

	for _, test := range tests {
		tst := test.ToDeserializeTest()
		t.Run(test.name, tst.Run)
	}
}

func TestBlockIssuerFeatureSyntacticValidation(t *testing.T) {
	bik1 := lo.PanicOnErr(lo.DropCount(
		iotago.Ed25519PublicKeyHashBlockIssuerKeyFromBytes(
			lo.PanicOnErr(hexutil.DecodeHex("0x00145d52e861cfe407e6f0c278f09ebd35ed7bcd766b7da2654e475ed4b05e0ddc")))))
	bik2 := lo.PanicOnErr(lo.DropCount(
		iotago.Ed25519PublicKeyHashBlockIssuerKeyFromBytes(
			lo.PanicOnErr(hexutil.DecodeHex("0x006f49dd17390fda4ec3b7c959496b4b9ac50428c47f0ffe445a94130547fbe519")))))
	bik3 := lo.PanicOnErr(lo.DropCount(
		iotago.Ed25519PublicKeyHashBlockIssuerKeyFromBytes(
			lo.PanicOnErr(hexutil.DecodeHex("0x009a224f3c94a5c281d984930216c20e1f4a79c3bad325cf92237f1dac1ff22b10")))))

	accountWithKeys := func(biks iotago.BlockIssuerKeys) *iotago.AccountOutput {
		return &iotago.AccountOutput{
			Amount: 100_000_000,
			UnlockConditions: iotago.AccountOutputUnlockConditions{
				&iotago.AddressUnlockCondition{
					Address: tpkg.RandAccountAddress(),
				},
			},
			ImmutableFeatures: iotago.AccountOutputImmFeatures{},
			Features: iotago.AccountOutputFeatures{
				&iotago.BlockIssuerFeature{
					ExpirySlot:      100,
					BlockIssuerKeys: biks,
				},
			},
		}
	}

	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - BlockIssuerFeature keys lexically ordered and unique",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *iotago.Transaction) {
				t.Outputs = iotago.TxEssenceOutputs{
					accountWithKeys(iotago.BlockIssuerKeys{
						bik1,
						bik2,
						bik3,
					}),
				}
				t.TransactionEssence.ContextInputs = append(t.TransactionEssence.ContextInputs, tpkg.RandCommitmentInput())
			},
			),
			Target: &iotago.SignedTransaction{},
		},
		{
			Name: "fail - BlockIssuerFeature keys lexically unordered",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *iotago.Transaction) {
				t.Outputs = iotago.TxEssenceOutputs{
					accountWithKeys(iotago.BlockIssuerKeys{
						bik2,
						bik1,
						bik3,
					}),
				}
				t.TransactionEssence.ContextInputs = append(t.TransactionEssence.ContextInputs, tpkg.RandCommitmentInput())
			}),
			Target:    &iotago.SignedTransaction{},
			SeriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			DeSeriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			Name: "fail - BlockIssuerFeature keys contains duplicates",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *iotago.Transaction) {
				t.Outputs = iotago.TxEssenceOutputs{
					accountWithKeys(iotago.BlockIssuerKeys{
						bik1,
						bik1,
						bik1,
						bik2,
					}),
				}
				t.TransactionEssence.ContextInputs = append(t.TransactionEssence.ContextInputs, tpkg.RandCommitmentInput())
			}),
			Target:    &iotago.SignedTransaction{},
			SeriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			DeSeriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			Name: "fail - BlockIssuerFeature keys below minimum",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *iotago.Transaction) {
				t.Outputs = iotago.TxEssenceOutputs{
					accountWithKeys(iotago.BlockIssuerKeys{}),
				}
				t.TransactionEssence.ContextInputs = append(t.TransactionEssence.ContextInputs, tpkg.RandCommitmentInput())
			}),
			Target:    &iotago.SignedTransaction{},
			SeriErr:   serializer.ErrArrayValidationMinElementsNotReached,
			DeSeriErr: serializer.ErrArrayValidationMinElementsNotReached,
		},
		{
			Name: "fail - BlockIssuerFeature keys exceeds maximum",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *iotago.Transaction) {
				t.Outputs = iotago.TxEssenceOutputs{
					accountWithKeys(tpkg.RandBlockIssuerKeys(iotago.MaxBlockIssuerKeysCount + 1)),
				}
				t.TransactionEssence.ContextInputs = append(t.TransactionEssence.ContextInputs, tpkg.RandCommitmentInput())
			}),
			Target:    &iotago.SignedTransaction{},
			SeriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			DeSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
