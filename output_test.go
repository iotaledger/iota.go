package iotago_test

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/tpkg"

	iotago "github.com/iotaledger/iota.go/v3"
)

func TestOutputTypeString(t *testing.T) {
	tests := []struct {
		outputType       iotago.OutputType
		outputTypeString string
	}{
		{iotago.OutputNFT, "NFTOutput"},
		{iotago.OutputTreasury, "TreasuryOutput"},
		{iotago.OutputBasic, "BasicOutput"},
		{iotago.OutputAlias, "AliasOutput"},
		{iotago.OutputFoundry, "FoundryOutput"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.outputType.String(), tt.outputTypeString)
	}
}

func TestOutputsDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - BasicOutput",
			source: &iotago.BasicOutput{
				Amount:       1337,
				NativeTokens: tpkg.RandSortNativeTokens(2),
				Conditions: iotago.UnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&iotago.TimelockUnlockCondition{UnixTime: 1337},
					&iotago.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						UnixTime:      4000,
					},
				},
				Features: iotago.Features{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
					&iotago.TagFeature{Tag: tpkg.RandBytes(32)},
				},
			},
			target: &iotago.BasicOutput{},
		},
		{
			name: "ok - AliasOutput",
			source: &iotago.AliasOutput{
				Amount:         1337,
				NativeTokens:   tpkg.RandSortNativeTokens(2),
				AliasID:        tpkg.RandAliasAddress().AliasID(),
				StateIndex:     10,
				StateMetadata:  []byte("hello world"),
				FoundryCounter: 1337,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.Features{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
				},
				ImmutableFeatures: iotago.Features{
					&iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
				},
			},
			target: &iotago.AliasOutput{},
		},
		{
			name: "ok - FoundryOutput",
			source: &iotago.FoundryOutput{
				Amount:       1337,
				NativeTokens: tpkg.RandSortNativeTokens(2),
				SerialNumber: 0,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetUint64(100),
					MeltedTokens:  big.NewInt(50),
					MaximumSupply: new(big.Int).SetUint64(1000),
				},
				Conditions: iotago.UnlockConditions{
					&iotago.ImmutableAliasUnlockCondition{Address: tpkg.RandAliasAddress()},
				},
				Features: iotago.Features{
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
				},
			},
			target: &iotago.FoundryOutput{},
		},
		{
			name: "ok - NFTOutput",
			source: &iotago.NFTOutput{
				Amount:       1337,
				NativeTokens: tpkg.RandSortNativeTokens(2),
				NFTID:        tpkg.Rand32ByteArray(),
				Conditions: iotago.UnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&iotago.TimelockUnlockCondition{UnixTime: 1337},
					&iotago.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						UnixTime:      4000,
					},
				},
				Features: iotago.Features{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
					&iotago.TagFeature{Tag: tpkg.RandBytes(32)},
				},
				ImmutableFeatures: iotago.Features{
					&iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(10)},
				},
			},
			target: &iotago.NFTOutput{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

type fieldMutations map[string]interface{}

func copyObject(t *testing.T, source serializer.Serializable, mutations fieldMutations, deSeriCtx interface{}) serializer.Serializable {
	srcBytes, err := source.Serialize(serializer.DeSeriModeNoValidation, deSeriCtx)
	require.NoError(t, err)

	ptrToCpyOfSrc := reflect.New(reflect.ValueOf(source).Elem().Type())

	cpySeri := ptrToCpyOfSrc.Interface().(serializer.Serializable)
	_, err = cpySeri.Deserialize(srcBytes, serializer.DeSeriModeNoValidation, deSeriCtx)
	require.NoError(t, err)

	for fieldName, newVal := range mutations {
		ptrToCpyOfSrc.Elem().FieldByName(fieldName).Set(reflect.ValueOf(newVal))
	}

	return cpySeri
}

func TestOutputsSyntacticalDepositAmount(t *testing.T) {
	nonZeroCostParas := &iotago.ProtocolParameters{
		RentStructure: iotago.RentStructure{
			VByteCost:    1,
			VBFactorData: iotago.VByteCostFactorData,
			VBFactorKey:  iotago.VByteCostFactorKey,
		},
		TokenSupply: tpkg.TestTokenSupply,
	}

	tests := []struct {
		name       string
		protoParas *iotago.ProtocolParameters
		outputs    iotago.Outputs
		wantErr    error
	}{
		{
			name:       "ok",
			protoParas: tpkg.TestProtoParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount:     tpkg.TestTokenSupply,
					Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()}},
				},
			},
			wantErr: nil,
		},
		{
			name:       "ok - state rent covered",
			protoParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount:     426, // min amount
					Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()}},
				},
			},
			wantErr: nil,
		},
		{
			name:       "ok - storage deposit return",
			protoParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				// min 444
				&iotago.BasicOutput{
					Amount: 1000,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAliasAddress(),
							Amount:        566, // 1000 - 444
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:       "fail - storage deposit return less than min storage deposit",
			protoParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: 1000,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAliasAddress(),
							Amount:        413, // off by 1
						},
					},
				},
			},
			wantErr: iotago.ErrStorageDepositLessThanMinReturnOutputStorageDeposit,
		},
		{
			name:       "fail - storage deposit more than target output deposit",
			protoParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAliasAddress(),
							// off by one from the deposit
							Amount: OneMi + 1,
						},
					},
				},
			},
			wantErr: iotago.ErrStorageDepositExceedsTargetOutputDeposit,
		},
		{
			name:       "fail - state rent not covered",
			protoParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
				},
			},
			wantErr: iotago.ErrVByteRentNotCovered,
		},
		{
			name:       "fail - zero deposit",
			protoParas: tpkg.TestProtoParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: 0,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrDepositAmountMustBeGreaterThanZero,
		},
		{
			name:       "fail - more than total supply on single output",
			protoParas: tpkg.TestProtoParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: tpkg.TestTokenSupply + 1,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrOutputDepositsMoreThanTotalSupply,
		},
		{
			name:       "fail - sum more than total supply over multiple outputs",
			protoParas: tpkg.TestProtoParas,
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: tpkg.TestTokenSupply - 1,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.BasicOutput{
					Amount: tpkg.TestTokenSupply - 1,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrOutputsSumExceedsTotalSupply,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.OutputsSyntacticalDepositAmount(tt.protoParas)
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticalNativeTokensCount(t *testing.T) {
	tests := []struct {
		name    string
		outputs iotago.Outputs
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(10),
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - sum more than max native tokens count",
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(50),
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(50),
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrMaxNativeTokensCountExceeded,
		},
		{
			name: "fail - native token with zero amount",
			outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: 1,
					NativeTokens: iotago.NativeTokens{
						&iotago.NativeToken{
							ID:     iotago.NativeTokenID{},
							Amount: big.NewInt(0),
						},
					},
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrNativeTokenAmountLessThanEqualZero,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.OutputsSyntacticalNativeTokens()
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticalAlias(t *testing.T) {
	type args struct {
		txID *iotago.TransactionID
	}
	tests := []struct {
		name    string
		outputs iotago.Outputs
		wantErr error
	}{
		{
			name: "ok - empty state",
			outputs: iotago.Outputs{
				&iotago.AliasOutput{
					Amount:         OneMi,
					AliasID:        iotago.AliasID{},
					StateIndex:     0,
					FoundryCounter: 0,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non empty state",
			outputs: iotago.Outputs{
				&iotago.AliasOutput{
					Amount:         OneMi,
					AliasID:        tpkg.Rand32ByteArray(),
					StateIndex:     10,
					FoundryCounter: 1337,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state index non zero on empty alias ID",
			outputs: iotago.Outputs{
				&iotago.AliasOutput{
					Amount:         OneMi,
					AliasID:        iotago.AliasID{},
					StateIndex:     1,
					FoundryCounter: 0,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
				},
			},
			wantErr: iotago.ErrAliasOutputNonEmptyState,
		},
		{
			name: "fail - foundry counter non zero on empty alias ID",
			outputs: iotago.Outputs{
				&iotago.AliasOutput{
					Amount:         OneMi,
					AliasID:        iotago.AliasID{},
					StateIndex:     0,
					FoundryCounter: 1,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
				},
			},
			wantErr: iotago.ErrAliasOutputNonEmptyState,
		},
		{
			name: "fail - cyclic state controller",
			outputs: iotago.Outputs{
				func() *iotago.AliasOutput {
					aliasID := iotago.AliasID(tpkg.Rand32ByteArray())
					return &iotago.AliasOutput{
						Amount:         OneMi,
						AliasID:        aliasID,
						StateIndex:     10,
						FoundryCounter: 1337,
						Conditions: iotago.UnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: aliasID.ToAddress()},
							&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAliasOutputCyclicAddress,
		},
		{
			name: "fail - cyclic governance controller",
			outputs: iotago.Outputs{
				func() *iotago.AliasOutput {
					aliasID := iotago.AliasID(tpkg.Rand32ByteArray())
					return &iotago.AliasOutput{
						Amount:         OneMi,
						AliasID:        aliasID,
						StateIndex:     10,
						FoundryCounter: 1337,
						Conditions: iotago.UnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAliasAddress()},
							&iotago.GovernorAddressUnlockCondition{Address: aliasID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAliasOutputCyclicAddress,
		},
	}
	valFunc := iotago.OutputsSyntacticalAlias()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				var runErr error
				for index, output := range tt.outputs {
					if err := valFunc(index, output); err != nil {
						runErr = err
					}
				}
				require.ErrorIs(t, runErr, tt.wantErr)
			})
		})
	}
}

func TestOutputsSyntacticalFoundry(t *testing.T) {
	tests := []struct {
		name    string
		outputs iotago.Outputs
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(5),
						MeltedTokens:  big.NewInt(2),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
					Features: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - minted and max supply same",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(10),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
					Features: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid maximum supply",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(5),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(0),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
					Features: nil,
				},
			},
			wantErr: iotago.ErrSimpleTokenSchemeInvalidMaximumSupply,
		},
		{
			name: "fail - minted less than melted",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(5),
						MeltedTokens:  big.NewInt(10),
						MaximumSupply: new(big.Int).SetUint64(100),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
					Features: nil,
				},
			},
			wantErr: iotago.ErrSimpleTokenSchemeInvalidMintedMeltedTokens,
		},
		{
			name: "fail - minted melted delta is bigger than maximum supply",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(50),
						MeltedTokens:  big.NewInt(20),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAliasAddress()},
					},
					Features: nil,
				},
			},
			wantErr: iotago.ErrSimpleTokenSchemeInvalidMintedMeltedTokens,
		},
	}
	valFunc := iotago.OutputsSyntacticalFoundry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				var runErr error
				for index, output := range tt.outputs {
					if err := valFunc(index, output); err != nil {
						runErr = err
					}
				}
				require.ErrorIs(t, runErr, tt.wantErr)
			})
		})
	}
}

func TestOutputsSyntacticalNFT(t *testing.T) {
	tests := []struct {
		name    string
		outputs iotago.Outputs
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs{
				&iotago.NFTOutput{
					Amount: OneMi,
					NFTID:  iotago.NFTID{},
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - cyclic",
			outputs: iotago.Outputs{
				func() *iotago.NFTOutput {
					nftID := iotago.NFTID(tpkg.Rand32ByteArray())
					return &iotago.NFTOutput{
						Amount: OneMi,
						NFTID:  nftID,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: nftID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrNFTOutputCyclicAddress,
		},
	}
	valFunc := iotago.OutputsSyntacticalNFT()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				var runErr error
				for index, output := range tt.outputs {
					if err := valFunc(index, output); err != nil {
						runErr = err
					}
				}
				require.ErrorIs(t, runErr, tt.wantErr)
			})
		})
	}
}

func TestTransIndepIdentOutput_UnlockableBy(t *testing.T) {
	type test struct {
		name                  string
		output                iotago.TransIndepIdentOutput
		targetIdent           iotago.Address
		identCanUnlockInstead iotago.Address
		extParas              *iotago.ExternalUnlockParameters
		canUnlock             bool
	}
	tests := []test{
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can unlock - target is source (no unlock conditions)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
					},
				},
				targetIdent: sourceIdent,
				extParas:    &iotago.ExternalUnlockParameters{},
				canUnlock:   true,
			}
		}(),
		func() test {
			return test{
				name: "can not unlock - target is not source (no unlock conditions)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				targetIdent: tpkg.RandEd25519Address(),
				extParas:    &iotago.ExternalUnlockParameters{},
				canUnlock:   false,
			}
		}(),
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can unlock - output not expired for source ident (unix expiration)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							UnixTime:      10,
						},
					},
				},
				targetIdent:           sourceIdent,
				identCanUnlockInstead: nil,
				extParas:              &iotago.ExternalUnlockParameters{ConfUnix: 5},
				canUnlock:             true,
			}
		}(),
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			senderIdent := tpkg.RandEd25519Address()
			return test{
				name: "can not unlock - output expired for source ident (unix expiration)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: senderIdent,
							UnixTime:      5,
						},
					},
				},
				targetIdent:           sourceIdent,
				identCanUnlockInstead: senderIdent,
				extParas:              &iotago.ExternalUnlockParameters{ConfUnix: 10},
				canUnlock:             false,
			}
		}(),
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can unlock - expired unix timelock unlock condition",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.TimelockUnlockCondition{UnixTime: 5},
					},
				},
				targetIdent: sourceIdent,
				extParas:    &iotago.ExternalUnlockParameters{ConfUnix: 10},
				canUnlock:   true,
			}
		}(),
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can not unlock - not expired unix timelock unlock condition",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.TimelockUnlockCondition{UnixTime: 10},
					},
				},
				targetIdent: sourceIdent,
				extParas:    &iotago.ExternalUnlockParameters{ConfUnix: 5},
				canUnlock:   false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.canUnlock, tt.output.UnlockableBy(tt.targetIdent, tt.extParas))
				if tt.identCanUnlockInstead == nil {
					return
				}
				require.True(t, tt.output.UnlockableBy(tt.identCanUnlockInstead, tt.extParas))
			})
		})
	}
}

func TestAliasOutput_UnlockableBy(t *testing.T) {
	type test struct {
		name                  string
		current               iotago.TransDepIdentOutput
		next                  iotago.TransDepIdentOutput
		targetIdent           iotago.Address
		identCanUnlockInstead iotago.Address
		extParas              *iotago.ExternalUnlockParameters
		wantErr               error
		canUnlock             bool
	}
	tests := []test{
		func() test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()
			return test{
				name: "state ctrl can unlock - state index increase",
				current: &iotago.AliasOutput{
					Amount:       OneMi,
					NativeTokens: nil,
					StateIndex:   0,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AliasOutput{
					Amount:     OneMi,
					StateIndex: 1,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				targetIdent: stateCtrl,
				extParas:    &iotago.ExternalUnlockParameters{},
				canUnlock:   true,
			}
		}(),
		func() test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()
			return test{
				name: "state ctrl can not unlock - state index same",
				current: &iotago.AliasOutput{
					Amount:       OneMi,
					NativeTokens: nil,
					StateIndex:   0,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AliasOutput{
					Amount:     OneMi,
					StateIndex: 0,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				targetIdent:           stateCtrl,
				identCanUnlockInstead: govCtrl,
				extParas:              &iotago.ExternalUnlockParameters{},
				canUnlock:             false,
			}
		}(),
		func() test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()
			return test{
				name: "state ctrl can not unlock - transition destroy",
				current: &iotago.AliasOutput{
					Amount:       OneMi,
					NativeTokens: nil,
					StateIndex:   0,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next:                  nil,
				targetIdent:           stateCtrl,
				identCanUnlockInstead: govCtrl,
				extParas:              &iotago.ExternalUnlockParameters{},
				canUnlock:             false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				canUnlock, err := tt.current.UnlockableBy(tt.targetIdent, tt.next, tt.extParas)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)
					return
				}
				require.Equal(t, tt.canUnlock, canUnlock)
				if tt.identCanUnlockInstead == nil {
					return
				}
				canUnlockInstead, err := tt.current.UnlockableBy(tt.identCanUnlockInstead, tt.next, tt.extParas)
				require.NoError(t, err)
				require.True(t, canUnlockInstead)
			})
		})
	}
}
