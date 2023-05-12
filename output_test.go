package iotago_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestOutputTypeString(t *testing.T) {
	tests := []struct {
		outputType       iotago.OutputType
		outputTypeString string
	}{
		{iotago.OutputNFT, "NFTOutput"},
		{iotago.OutputTreasury, "TreasuryOutput"},
		{iotago.OutputBasic, "BasicOutput"},
		{iotago.OutputAccount, "AccountOutput"},
		{iotago.OutputFoundry, "FoundryOutput"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.outputType.String(), tt.outputTypeString)
	}
}

func TestOutputIDString(t *testing.T) {
	tests := []struct {
		outputID         iotago.OutputID
		outputTypeString string
	}{
		{outputID: iotago.OutputIDFromTransactionIDAndIndex(lo.PanicOnErr(iotago.IdentifierFromHexString("0xc8ed3cbe4acb99aeb94515ad89a6228f3f5d8f82dec429df135adafcea639416")), 1), outputTypeString: "OutputID(0xc8ed3cbe4acb99aeb94515ad89a6228f3f5d8f82dec429df135adafcea639416:1)"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.outputID.String(), tt.outputTypeString)
	}
}

func TestOutputsDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - BasicOutput",
			source: &iotago.BasicOutput{
				Amount:       1337,
				NativeTokens: tpkg.RandSortNativeTokens(2),
				Conditions: iotago.BasicOutputUnlockConditions{
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
				Features: iotago.BasicOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
					&iotago.TagFeature{Tag: tpkg.RandBytes(32)},
				},
			},
			target: &iotago.BasicOutput{},
		},
		{
			name: "ok - AccountOutput",
			source: &iotago.AccountOutput{
				Amount:         1337,
				NativeTokens:   tpkg.RandSortNativeTokens(2),
				AccountID:      tpkg.RandAccountAddress().AccountID(),
				StateIndex:     10,
				StateMetadata:  []byte("hello world"),
				FoundryCounter: 1337,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
				},
			},
			target: &iotago.AccountOutput{},
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
				Conditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
				},
				Features: iotago.FoundryOutputFeatures{
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
				Conditions: iotago.NFTOutputUnlockConditions{
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
				Features: iotago.NFTOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
					&iotago.TagFeature{Tag: tpkg.RandBytes(32)},
				},
				ImmutableFeatures: iotago.NFTOutputImmFeatures{
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

func TestOutputsSyntacticalDepositAmount(t *testing.T) {
	nonZeroCostParams := &iotago.ProtocolParameters{
		RentStructure: iotago.RentStructure{
			VByteCost:    100,
			VBFactorData: iotago.VByteCostFactorData,
			VBFactorKey:  iotago.VByteCostFactorKey,
		},
		TokenSupply: tpkg.TestTokenSupply,
	}

	tests := []struct {
		name        string
		protoParams *iotago.ProtocolParameters
		outputs     iotago.Outputs[iotago.Output]
		wantErr     error
	}{
		{
			name:        "ok",
			protoParams: tpkg.TestProtoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:     tpkg.TestTokenSupply,
					Conditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()}},
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - state rent covered",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:     42600, // min amount
					Conditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()}},
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - storage deposit return",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				// min 46800
				&iotago.BasicOutput{
					Amount: 100000,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							Amount:        42600,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:        "fail - storage deposit return less than min storage deposit",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 100000,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							Amount:        42600 - 1, // off by 1
						},
					},
				},
			},
			wantErr: iotago.ErrStorageDepositLessThanMinReturnOutputStorageDeposit,
		},
		{
			name:        "fail - storage deposit more than target output deposit",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							// off by one from the deposit
							Amount: OneMi + 1,
						},
					},
				},
			},
			wantErr: iotago.ErrStorageDepositExceedsTargetOutputDeposit,
		},
		{
			name:        "fail - state rent not covered",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 42600 - 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrVByteRentNotCovered,
		},
		{
			name:        "fail - zero deposit",
			protoParams: tpkg.TestProtoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 0,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrDepositAmountMustBeGreaterThanZero,
		},
		{
			name:        "fail - more than total supply on single output",
			protoParams: tpkg.TestProtoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: tpkg.TestTokenSupply + 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrOutputDepositsMoreThanTotalSupply,
		},
		{
			name:        "fail - sum more than total supply over multiple outputs",
			protoParams: tpkg.TestProtoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: tpkg.TestTokenSupply - 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.BasicOutput{
					Amount: tpkg.TestTokenSupply - 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrOutputsSumExceedsTotalSupply,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.OutputsSyntacticalDepositAmount(tt.protoParams)
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr, tt.name)
		})
	}
}

func TestOutputsSyntacticalExpirationAndTimelock(t *testing.T) {
	tests := []struct {
		name    string
		outputs iotago.TxEssenceOutputs
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.TxEssenceOutputs{
				&iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							UnixTime:      1337,
						},
					},
				},
				&iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.TimelockUnlockCondition{
							UnixTime: 1337,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - zero expiration time",
			outputs: iotago.TxEssenceOutputs{
				&iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							UnixTime:      0,
						},
					},
				},
			},
			wantErr: iotago.ErrExpirationConditionZero,
		},
		{
			name: "fail - zero timelock time",
			outputs: iotago.TxEssenceOutputs{
				&iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.TimelockUnlockCondition{
							UnixTime: 0,
						},
					},
				},
			},
			wantErr: iotago.ErrTimelockConditionZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.OutputsSyntacticalExpirationAndTimelock()
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
		outputs iotago.Outputs[iotago.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(10),
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - sum more than max native tokens count",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(50),
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(50),
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrMaxNativeTokensCountExceeded,
		},
		{
			name: "fail - native token with zero amount",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 1,
					NativeTokens: iotago.NativeTokens{
						&iotago.NativeToken{
							ID:     iotago.NativeTokenID{},
							Amount: big.NewInt(0),
						},
					},
					Conditions: iotago.BasicOutputUnlockConditions{
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

func TestOutputsSyntacticalAccount(t *testing.T) {
	tests := []struct {
		name    string
		outputs iotago.Outputs[iotago.Output]
		wantErr error
	}{
		{
			name: "ok - empty state",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneMi,
					AccountID:      iotago.AccountID{},
					StateIndex:     0,
					FoundryCounter: 0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non empty state",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneMi,
					AccountID:      tpkg.Rand32ByteArray(),
					StateIndex:     10,
					FoundryCounter: 1337,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state index non zero on empty account ID",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneMi,
					AccountID:      iotago.AccountID{},
					StateIndex:     1,
					FoundryCounter: 0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrAccountOutputNonEmptyState,
		},
		{
			name: "fail - foundry counter non zero on empty account ID",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneMi,
					AccountID:      iotago.AccountID{},
					StateIndex:     0,
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrAccountOutputNonEmptyState,
		},
		{
			name: "fail - cyclic state controller",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AccountOutput {
					accountID := iotago.AccountID(tpkg.Rand32ByteArray())
					return &iotago.AccountOutput{
						Amount:         OneMi,
						AccountID:      accountID,
						StateIndex:     10,
						FoundryCounter: 1337,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: accountID.ToAddress()},
							&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAccountOutputCyclicAddress,
		},
		{
			name: "fail - cyclic governance controller",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AccountOutput {
					accountID := iotago.AccountID(tpkg.Rand32ByteArray())
					return &iotago.AccountOutput{
						Amount:         OneMi,
						AccountID:      accountID,
						StateIndex:     10,
						FoundryCounter: 1337,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
							&iotago.GovernorAddressUnlockCondition{Address: accountID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAccountOutputCyclicAddress,
		},
	}
	valFunc := iotago.OutputsSyntacticalAccount()
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
		outputs iotago.Outputs[iotago.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(5),
						MeltedTokens:  big.NewInt(2),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - minted and max supply same",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(10),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid maximum supply",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(5),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(0),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: iotago.ErrSimpleTokenSchemeInvalidMaximumSupply,
		},
		{
			name: "fail - minted less than melted",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(5),
						MeltedTokens:  big.NewInt(10),
						MaximumSupply: new(big.Int).SetUint64(100),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: iotago.ErrSimpleTokenSchemeInvalidMintedMeltedTokens,
		},
		{
			name: "fail - minted melted delta is bigger than maximum supply",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.FoundryOutput{
					Amount:       1337,
					NativeTokens: nil,
					SerialNumber: 5,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(50),
						MeltedTokens:  big.NewInt(20),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
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
		outputs iotago.Outputs[iotago.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.NFTOutput{
					Amount: OneMi,
					NFTID:  iotago.NFTID{},
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - cyclic",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.NFTOutput {
					nftID := iotago.NFTID(tpkg.Rand32ByteArray())
					return &iotago.NFTOutput{
						Amount: OneMi,
						NFTID:  nftID,
						Conditions: iotago.NFTOutputUnlockConditions{
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
		extParams             *iotago.ExternalUnlockParameters
		canUnlock             bool
	}
	tests := []test{
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can unlock - target is source (no unlock conditions)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
					},
				},
				targetIdent: sourceIdent,
				extParams:   &iotago.ExternalUnlockParameters{},
				canUnlock:   true,
			}
		}(),
		func() test {
			return test{
				name: "can not unlock - target is not source (no unlock conditions)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				targetIdent: tpkg.RandEd25519Address(),
				extParams:   &iotago.ExternalUnlockParameters{},
				canUnlock:   false,
			}
		}(),
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can unlock - output not expired for source ident (unix expiration)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							UnixTime:      10,
						},
					},
				},
				targetIdent:           sourceIdent,
				identCanUnlockInstead: nil,
				extParams:             &iotago.ExternalUnlockParameters{ConfUnix: 5},
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
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: senderIdent,
							UnixTime:      5,
						},
					},
				},
				targetIdent:           sourceIdent,
				identCanUnlockInstead: senderIdent,
				extParams:             &iotago.ExternalUnlockParameters{ConfUnix: 10},
				canUnlock:             false,
			}
		}(),
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can unlock - expired unix timelock unlock condition",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.TimelockUnlockCondition{UnixTime: 5},
					},
				},
				targetIdent: sourceIdent,
				extParams:   &iotago.ExternalUnlockParameters{ConfUnix: 10},
				canUnlock:   true,
			}
		}(),
		func() test {
			sourceIdent := tpkg.RandEd25519Address()
			return test{
				name: "can not unlock - not expired unix timelock unlock condition",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: sourceIdent},
						&iotago.TimelockUnlockCondition{UnixTime: 10},
					},
				},
				targetIdent: sourceIdent,
				extParams:   &iotago.ExternalUnlockParameters{ConfUnix: 5},
				canUnlock:   false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.canUnlock, tt.output.UnlockableBy(tt.targetIdent, tt.extParams))
				if tt.identCanUnlockInstead == nil {
					return
				}
				require.True(t, tt.output.UnlockableBy(tt.identCanUnlockInstead, tt.extParams))
			})
		})
	}
}

func TestAccountOutput_UnlockableBy(t *testing.T) {
	type test struct {
		name                  string
		current               iotago.TransDepIdentOutput
		next                  iotago.TransDepIdentOutput
		targetIdent           iotago.Address
		identCanUnlockInstead iotago.Address
		extParams             *iotago.ExternalUnlockParameters
		wantErr               error
		canUnlock             bool
	}
	tests := []test{
		func() test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()
			return test{
				name: "state ctrl can unlock - state index increase",
				current: &iotago.AccountOutput{
					Amount:       OneMi,
					NativeTokens: nil,
					StateIndex:   0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AccountOutput{
					Amount:     OneMi,
					StateIndex: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				targetIdent: stateCtrl,
				extParams:   &iotago.ExternalUnlockParameters{},
				canUnlock:   true,
			}
		}(),
		func() test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()
			return test{
				name: "state ctrl can not unlock - state index same",
				current: &iotago.AccountOutput{
					Amount:       OneMi,
					NativeTokens: nil,
					StateIndex:   0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AccountOutput{
					Amount:     OneMi,
					StateIndex: 0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				targetIdent:           stateCtrl,
				identCanUnlockInstead: govCtrl,
				extParams:             &iotago.ExternalUnlockParameters{},
				canUnlock:             false,
			}
		}(),
		func() test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()
			return test{
				name: "state ctrl can not unlock - transition destroy",
				current: &iotago.AccountOutput{
					Amount:       OneMi,
					NativeTokens: nil,
					StateIndex:   0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next:                  nil,
				targetIdent:           stateCtrl,
				identCanUnlockInstead: govCtrl,
				extParams:             &iotago.ExternalUnlockParameters{},
				canUnlock:             false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				canUnlock, err := tt.current.UnlockableBy(tt.targetIdent, tt.next, tt.extParams)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)
					return
				}
				require.Equal(t, tt.canUnlock, canUnlock)
				if tt.identCanUnlockInstead == nil {
					return
				}
				canUnlockInstead, err := tt.current.UnlockableBy(tt.identCanUnlockInstead, tt.next, tt.extParams)
				require.NoError(t, err)
				require.True(t, canUnlockInstead)
			})
		})
	}
}
