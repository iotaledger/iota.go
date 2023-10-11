//nolint:scopelint
package iotago_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestOutputTypeString(t *testing.T) {
	tests := []struct {
		outputType       iotago.OutputType
		outputTypeString string
	}{
		{iotago.OutputBasic, "BasicOutput"},
		{iotago.OutputAccount, "AccountOutput"},
		{iotago.OutputFoundry, "FoundryOutput"},
		{iotago.OutputNFT, "NFTOutput"},
		{iotago.OutputAnchor, "AnchorOutput"},
		{iotago.OutputDelegation, "DelegationOutput"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.outputType.String(), tt.outputTypeString)
	}
}

func TestOutputsDeSerialize(t *testing.T) {
	emptyAccountAddress := iotago.AccountAddress(iotago.EmptyAccountID)

	tests := []deSerializeTest{
		{
			name: "ok - BasicOutput",
			source: &iotago.BasicOutput{
				Amount: 1337,
				Mana:   500,
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
					tpkg.RandNativeTokenFeature(),
				},
			},
			target: &iotago.BasicOutput{},
		},
		{
			name: "ok - AccountOutput",
			source: &iotago.AccountOutput{
				Amount:         1337,
				Mana:           500,
				AccountID:      tpkg.RandAccountAddress().AccountID(),
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
			name: "ok - AnchorOutput",
			source: &iotago.AnchorOutput{
				Amount:        1337,
				Mana:          500,
				AnchorID:      tpkg.RandAnchorAddress().AnchorID(),
				StateIndex:    10,
				StateMetadata: []byte("hello world"),
				Conditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AnchorOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Data: tpkg.RandBytes(100)},
				},
				ImmutableFeatures: iotago.AnchorOutputImmFeatures{
					&iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
				},
			},
			target: &iotago.AnchorOutput{},
		},
		{
			name: "ok - FoundryOutput",
			source: &iotago.FoundryOutput{
				Amount:       1337,
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
				Amount: 1337,
				Mana:   500,
				NFTID:  tpkg.Rand32ByteArray(),
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
				Amount:           1337,
				DelegatedAmount:  1337,
				DelegationID:     tpkg.Rand32ByteArray(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       iotago.EpochIndex(32),
				EndEpoch:         iotago.EpochIndex(37),
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			target: &iotago.DelegationOutput{},
		},
		{
			name: "fail - Delegation Output contains Implicit Account Creation Address",
			source: &iotago.DelegationOutput{
				Amount:           1337,
				DelegatedAmount:  1337,
				DelegationID:     tpkg.Rand32ByteArray(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       iotago.EpochIndex(32),
				EndEpoch:         iotago.EpochIndex(37),
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			target:  &iotago.DelegationOutput{},
			seriErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - NFT Output contains Implicit Account Creation Address",
			source: &iotago.NFTOutput{
				Conditions: iotago.NFTOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			target:  &iotago.NFTOutput{},
			seriErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Account Output contains Implicit Account Creation Address as State Controller",
			source: &iotago.AccountOutput{
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			target:  &iotago.AccountOutput{},
			seriErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Account Output contains Implicit Account Creation Address as Governor",
			source: &iotago.AccountOutput{
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			target:  &iotago.AccountOutput{},
			seriErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Delegation Output contains empty validator address",
			source: &iotago.DelegationOutput{
				ValidatorAddress: &emptyAccountAddress,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			target:  &iotago.DelegationOutput{},
			seriErr: iotago.ErrDelegationValidatorAddressEmpty,
		},
		{
			name: "fail - Anchor Output contains Implicit Account Creation Address as State Controller",
			source: &iotago.AnchorOutput{
				Conditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			target:  &iotago.AnchorOutput{},
			seriErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Anchor Output contains Implicit Account Creation Address as Governor",
			source: &iotago.AnchorOutput{
				Conditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			target:  &iotago.AnchorOutput{},
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

	var minAmount iotago.BaseToken = 14100

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
					Amount:     minAmount, // min amount
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
							Amount:        minAmount, // min amount
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
							Amount:        minAmount - 1, // off by 1
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
					Amount: minAmount - 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrStorageDepositNotCovered,
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
			valFunc := iotago.OutputsSyntacticalDepositAmount(tt.protoParams, iotago.NewRentStructure(tt.protoParams.RentParameters()))
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
					Amount: 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				},
				&iotago.BasicOutput{
					Amount: 1,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - native token with zero amount",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 1,
					Features: iotago.BasicOutputFeatures{
						&iotago.NativeTokenFeature{
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
			name: "fail - foundry counter non zero on empty account ID",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneMi,
					AccountID:      iotago.AccountID{},
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrAnchorOutputNonEmptyState,
		},
		{
			name: "fail - cyclic",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AccountOutput {
					accountID := iotago.AccountID(tpkg.Rand32ByteArray())

					return &iotago.AccountOutput{
						Amount:         OneMi,
						AccountID:      accountID,
						FoundryCounter: 1337,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: accountID.ToAddress()},
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
	emptyAccountAddress := iotago.AccountAddress{}

	tests := []struct {
		name    string
		outputs iotago.Outputs[iotago.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.DelegationOutput{
					Amount:           OneMi,
					DelegatedAmount:  OneMi,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
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
					Amount:           OneMi,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: &emptyAccountAddress,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrDelegationValidatorAddressZeroed,
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

func TestOutputsSyntacticalAnchor(t *testing.T) {
	tests := []struct {
		name    string
		outputs iotago.Outputs[iotago.Output]
		wantErr error
	}{
		{
			name: "ok - empty state",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AnchorOutput{
					Amount:        OneMi,
					AnchorID:      iotago.AnchorID{},
					StateIndex:    0,
					StateMetadata: []byte{},
					Conditions: iotago.AnchorOutputUnlockConditions{
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
				&iotago.AnchorOutput{
					Amount:        OneMi,
					AnchorID:      tpkg.Rand32ByteArray(),
					StateIndex:    10,
					StateMetadata: []byte{},
					Conditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - cyclic state controller",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AnchorOutput {
					anchorID := iotago.AnchorID(tpkg.Rand32ByteArray())

					return &iotago.AnchorOutput{
						Amount:     OneMi,
						AnchorID:   anchorID,
						StateIndex: 10,
						Conditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: anchorID.ToAddress()},
							&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAnchorOutputCyclicAddress,
		},
		{
			name: "fail - cyclic governance controller",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AnchorOutput {
					anchorID := iotago.AnchorID(tpkg.Rand32ByteArray())

					return &iotago.AnchorOutput{
						Amount:     OneMi,
						AnchorID:   anchorID,
						StateIndex: 10,
						Conditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAccountAddress()},
							&iotago.GovernorAddressUnlockCondition{Address: anchorID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAnchorOutputCyclicAddress,
		},
	}
	valFunc := iotago.OutputsSyntacticalAnchor()
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
	tests := []*test{
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			return &test{
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
		func() *test {
			return &test{
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
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			return &test{
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
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			returnIdent := tpkg.RandEd25519Address()
			return &test{
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
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			returnIdent := tpkg.RandEd25519Address()
			return &test{
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
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			returnIdent := tpkg.RandEd25519Address()
			return &test{
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
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			return &test{
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
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			return &test{
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
	tests := []*test{
		func() *test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()

			return &test{
				name: "state ctrl can unlock - state index increase",
				current: &iotago.AnchorOutput{
					Amount:     OneMi,
					StateIndex: 0,
					Conditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AnchorOutput{
					Amount:     OneMi,
					StateIndex: 1,
					Conditions: iotago.AnchorOutputUnlockConditions{
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
		func() *test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()

			return &test{
				name: "state ctrl can not unlock - state index same",
				current: &iotago.AnchorOutput{
					Amount:     OneMi,
					StateIndex: 0,
					Conditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AnchorOutput{
					Amount:     OneMi,
					StateIndex: 0,
					Conditions: iotago.AnchorOutputUnlockConditions{
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
		func() *test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()

			return &test{
				name: "state ctrl can not unlock - transition destroy",
				current: &iotago.AnchorOutput{
					Amount:     OneMi,
					StateIndex: 0,
					Conditions: iotago.AnchorOutputUnlockConditions{
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
