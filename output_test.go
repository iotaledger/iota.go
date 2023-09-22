//nolint:scopelint
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
func TestOutputsCommitment(t *testing.T) {
	outputs1 := iotago.Outputs[iotago.Output]{
		&iotago.BasicOutput{
			Amount: 10,
			Conditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
			},
		},
		&iotago.BasicOutput{
			Amount: 10,
			Conditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
			},
		},
	}

	outputs2 := iotago.Outputs[iotago.Output]{
		&iotago.BasicOutput{
			Amount: 11,
			Conditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
			},
		},
		&iotago.BasicOutput{
			Amount: 11,
			Conditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
			},
		},
	}

	require.NotEqual(t, outputs1.MustCommitment(tpkg.TestAPI), outputs2.MustCommitment(tpkg.TestAPI), "commitment for different Outputs must be different")

}

func TestOutputIDString(t *testing.T) {
	tests := []struct {
		outputID         iotago.OutputID
		outputTypeString string
	}{
		{outputID: iotago.OutputIDFromTransactionIDAndIndex(lo.PanicOnErr(iotago.TransactionIDFromHexString("0xbaadf00ddeadbeefc8ed3cbe4acb99aeb94515ad89a6228f3f5d8f82dec429df135adafcea639416")), 1), outputTypeString: "OutputID(0xbaadf00ddeadbeefc8ed3cbe4acb99aeb94515ad89a6228f3f5d8f82dec429df135adafcea639416:1)"},
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
				Mana:         500,
				NativeTokens: tpkg.RandSortNativeTokens(2),
				Conditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&iotago.TimelockUnlockCondition{SlotIndex: 1337},
					&iotago.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						SlotIndex:     4000,
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
				Mana:           500,
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
				ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
			},
			target: &iotago.FoundryOutput{},
		},
		{
			name: "ok - NFTOutput",
			source: &iotago.NFTOutput{
				Amount:       1337,
				Mana:         500,
				NativeTokens: tpkg.RandSortNativeTokens(2),
				NFTID:        tpkg.Rand32ByteArray(),
				Conditions: iotago.NFTOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&iotago.TimelockUnlockCondition{SlotIndex: 1337},
					&iotago.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						SlotIndex:     4000,
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
		{
			name: "ok - DelegationOutput",
			source: &iotago.DelegationOutput{
				Amount:          1337,
				DelegatedAmount: 1337,
				DelegationID:    tpkg.Rand32ByteArray(),
				ValidatorID:     tpkg.RandAccountID(),
				StartEpoch:      iotago.EpochIndex(32),
				EndEpoch:        iotago.EpochIndex(37),
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			target: &iotago.DelegationOutput{},
		},
		{
			name: "fail - Delegation Output contains Implicit Account Creation Address",
			source: &iotago.DelegationOutput{
				Amount:          1337,
				DelegatedAmount: 1337,
				DelegationID:    tpkg.Rand32ByteArray(),
				ValidatorID:     tpkg.RandAccountID(),
				StartEpoch:      iotago.EpochIndex(32),
				EndEpoch:        iotago.EpochIndex(37),
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			target:  &iotago.DelegationOutput{},
			seriErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestOutputsSyntacticalDepositAmount(t *testing.T) {
	nonZeroCostParams := iotago.NewV3ProtocolParameters(
		iotago.WithSupplyOptions(tpkg.TestTokenSupply, 100, 1, 10, 10, 10, 10),
	)

	tests := []struct {
		name        string
		protoParams iotago.ProtocolParameters
		outputs     iotago.Outputs[iotago.Output]
		wantErr     error
	}{
		{
			name:        "ok",
			protoParams: tpkg.TestAPI.ProtocolParameters(),
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:     tpkg.TestTokenSupply,
					Conditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()}},
					Mana:       500,
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - state rent covered",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:     52200, // min amount
					Conditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()}},
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - storage deposit return",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 100000,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							Amount:        52200, // amount
						},
					},
					Mana: 500,
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
							Amount:        52200 - 1, // off by 1
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
					Mana: 500,
				},
			},
			wantErr: iotago.ErrStorageDepositExceedsTargetOutputAmount,
		},
		{
			name:        "fail - state rent not covered",
			protoParams: nonZeroCostParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 52200 - 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrVByteDepositNotCovered,
		},
		{
			name:        "fail - zero deposit",
			protoParams: tpkg.TestAPI.ProtocolParameters(),
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 0,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrAmountMustBeGreaterThanZero,
		},
		{
			name:        "fail - more than total supply on single output",
			protoParams: tpkg.TestAPI.ProtocolParameters(),
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: tpkg.TestTokenSupply + 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrOutputAmountMoreThanTotalSupply,
		},
		{
			name:        "fail - sum more than total supply over multiple outputs",
			protoParams: tpkg.TestAPI.ProtocolParameters(),
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
							SlotIndex:     1337,
						},
					},
				},
				&iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.TimelockUnlockCondition{
							SlotIndex: 1337,
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
							SlotIndex:     0,
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
							SlotIndex: 0,
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

func TestOutputsSyntacticaDelegation(t *testing.T) {
	tests := []struct {
		name    string
		outputs iotago.Outputs[iotago.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.DelegationOutput{
					Amount:          OneMi,
					DelegatedAmount: OneMi,
					DelegationID:    iotago.EmptyDelegationID(),
					ValidatorID:     tpkg.RandAccountID(),
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - validator id zeroed",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.DelegationOutput{
					Amount:       OneMi,
					DelegationID: iotago.EmptyDelegationID(),
					ValidatorID:  iotago.EmptyAccountID(),
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrDelegationValidatorIDZeroed,
		},
	}
	valFunc := iotago.OutputsSyntacticalDelegation()
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
		name                string
		output              iotago.TransIndepIdentOutput
		targetIdent         iotago.Address
		commitmentInputTime iotago.SlotIndex
		minCommittableAge   iotago.SlotIndex
		maxCommittableAge   iotago.SlotIndex
		canUnlock           bool
	}
	tests := []test{
		func() test {
			receiverIdent := tpkg.RandEd25519Address()
			return test{
				name: "can unlock - target is source (no unlock conditions)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
					},
				},
				targetIdent:         receiverIdent,
				commitmentInputTime: iotago.SlotIndex(0),
				minCommittableAge:   iotago.SlotIndex(0),
				maxCommittableAge:   iotago.SlotIndex(0),
				canUnlock:           true,
			}
		}(),
		func() test {
			return test{
				name: "can not unlock - target is not source (no timelocks or expiration unlock conditions)",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				targetIdent:         tpkg.RandEd25519Address(),
				commitmentInputTime: iotago.SlotIndex(0),
				minCommittableAge:   iotago.SlotIndex(0),
				maxCommittableAge:   iotago.SlotIndex(0),
				canUnlock:           false,
			}
		}(),
		func() test {
			receiverIdent := tpkg.RandEd25519Address()
			return test{
				name: "expiration - receiver ident can unlock",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							SlotIndex:     26,
						},
					},
				},
				targetIdent:         receiverIdent,
				commitmentInputTime: iotago.SlotIndex(5),
				minCommittableAge:   iotago.SlotIndex(10),
				maxCommittableAge:   iotago.SlotIndex(20),
				canUnlock:           true,
			}
		}(),
		func() test {
			receiverIdent := tpkg.RandEd25519Address()
			returnIdent := tpkg.RandEd25519Address()
			return test{
				name: "expiration - receiver ident can not unlock",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: returnIdent,
							SlotIndex:     25,
						},
					},
				},
				targetIdent:         receiverIdent,
				commitmentInputTime: iotago.SlotIndex(5),
				minCommittableAge:   iotago.SlotIndex(10),
				maxCommittableAge:   iotago.SlotIndex(20),
				canUnlock:           false,
			}
		}(),
		func() test {
			receiverIdent := tpkg.RandEd25519Address()
			returnIdent := tpkg.RandEd25519Address()
			return test{
				name: "expiration - return ident can unlock",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: returnIdent,
							SlotIndex:     15,
						},
					},
				},
				targetIdent:         returnIdent,
				commitmentInputTime: iotago.SlotIndex(5),
				minCommittableAge:   iotago.SlotIndex(10),
				maxCommittableAge:   iotago.SlotIndex(20),
				canUnlock:           true,
			}
		}(),
		func() test {
			receiverIdent := tpkg.RandEd25519Address()
			returnIdent := tpkg.RandEd25519Address()
			return test{
				name: "expiration - return ident can not unlock",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: returnIdent,
							SlotIndex:     16,
						},
					},
				},
				targetIdent:         returnIdent,
				commitmentInputTime: iotago.SlotIndex(5),
				minCommittableAge:   iotago.SlotIndex(10),
				maxCommittableAge:   iotago.SlotIndex(20),
				canUnlock:           false,
			}
		}(),
		func() test {
			receiverIdent := tpkg.RandEd25519Address()
			return test{
				name: "timelock - expired timelock is unlockable",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.TimelockUnlockCondition{SlotIndex: 15},
					},
				},
				targetIdent:         receiverIdent,
				commitmentInputTime: iotago.SlotIndex(5),
				minCommittableAge:   iotago.SlotIndex(10),
				maxCommittableAge:   iotago.SlotIndex(20),
				canUnlock:           true,
			}
		}(),
		func() test {
			receiverIdent := tpkg.RandEd25519Address()
			return test{
				name: "timelock - non-expired timelock is not unlockable",
				output: &iotago.BasicOutput{
					Amount: OneMi,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.TimelockUnlockCondition{SlotIndex: 16},
					},
				},
				targetIdent:         receiverIdent,
				commitmentInputTime: iotago.SlotIndex(5),
				minCommittableAge:   iotago.SlotIndex(10),
				maxCommittableAge:   iotago.SlotIndex(20),
				canUnlock:           false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.canUnlock, tt.output.UnlockableBy(tt.targetIdent, tt.commitmentInputTime+tt.maxCommittableAge, tt.commitmentInputTime+tt.minCommittableAge))
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
		commitmentInputTime   iotago.SlotIndex
		minCommittableAge     iotago.SlotIndex
		maxCommittableAge     iotago.SlotIndex
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
				targetIdent:         stateCtrl,
				commitmentInputTime: iotago.SlotIndex(0),
				minCommittableAge:   iotago.SlotIndex(0),
				maxCommittableAge:   iotago.SlotIndex(0),
				canUnlock:           true,
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
				commitmentInputTime:   iotago.SlotIndex(0),
				minCommittableAge:     iotago.SlotIndex(0),
				maxCommittableAge:     iotago.SlotIndex(0),
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
				commitmentInputTime:   iotago.SlotIndex(0),
				minCommittableAge:     iotago.SlotIndex(0),
				maxCommittableAge:     iotago.SlotIndex(0),
				canUnlock:             false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				canUnlock, err := tt.current.UnlockableBy(tt.targetIdent, tt.next, tt.commitmentInputTime+tt.maxCommittableAge, tt.commitmentInputTime+tt.minCommittableAge)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)

					return
				}
				require.Equal(t, tt.canUnlock, canUnlock)
				if tt.identCanUnlockInstead == nil {
					return
				}
				canUnlockInstead, err := tt.current.UnlockableBy(tt.identCanUnlockInstead, tt.next, tt.commitmentInputTime+tt.maxCommittableAge, tt.commitmentInputTime+tt.minCommittableAge)
				require.NoError(t, err)
				require.True(t, canUnlockInstead)
			})
		})
	}
}
