//nolint:scopelint
package iotago_test

import (
	"fmt"
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
		{iotago.OutputAnchor, "AnchorOutput"},
		{iotago.OutputFoundry, "FoundryOutput"},
		{iotago.OutputNFT, "NFTOutput"},
		{iotago.OutputDelegation, "DelegationOutput"},
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
				Amount: 1337,
				Mana:   500,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&iotago.TimelockUnlockCondition{Slot: 1337},
					&iotago.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Slot:          4000,
					},
				},
				Features: iotago.BasicOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
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
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
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
				Amount:     1337,
				Mana:       500,
				AnchorID:   tpkg.RandAnchorAddress().AnchorID(),
				StateIndex: 10,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AnchorOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
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
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
				},
				Features: iotago.FoundryOutputFeatures{
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
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
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&iotago.TimelockUnlockCondition{Slot: 1337},
					&iotago.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Slot:          4000,
					},
				},
				Features: iotago.NFTOutputFeatures{
					&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
					&iotago.TagFeature{Tag: tpkg.RandBytes(32)},
				},
				ImmutableFeatures: iotago.NFTOutputImmFeatures{
					&iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(10)}},
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
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			target: &iotago.DelegationOutput{},
		},
		{
			name: "fail - BasicOutput contains duplicate unlock conditions",
			source: &iotago.BasicOutput{
				Amount: 1337,
				Mana:   500,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.TimelockUnlockCondition{Slot: 1337},
					&iotago.TimelockUnlockCondition{Slot: 1000},
				},
				Features: iotago.BasicOutputFeatures{},
			},
			target:    &iotago.BasicOutput{},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deSeriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - BasicOutput contains lexically unordered unlock conditions",
			source: &iotago.BasicOutput{
				Amount: 1337,
				Mana:   500,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					// Unlock Condition Type 0
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					// Unlock Condition Type 3
					&iotago.ExpirationUnlockCondition{Slot: 1000},
					// Unlock Condition Type 2
					&iotago.TimelockUnlockCondition{Slot: 1337},
				},
				Features: iotago.BasicOutputFeatures{},
			},
			target:    &iotago.BasicOutput{},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deSeriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - BasicOutput contains duplicate features",
			source: &iotago.BasicOutput{
				Amount: 1337,
				Mana:   500,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.BasicOutputFeatures{
					&iotago.TagFeature{
						Tag: tpkg.RandBytes(3),
					},
					&iotago.TagFeature{
						Tag: tpkg.RandBytes(6),
					},
				},
			},
			target:    &iotago.BasicOutput{},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deSeriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - BasicOutput contains lexically unordered features",
			source: &iotago.BasicOutput{
				Amount: 1337,
				Mana:   500,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.BasicOutputFeatures{
					// Feature Type 4
					&iotago.TagFeature{
						Tag: tpkg.RandBytes(3),
					},
					// Feature Type 0
					&iotago.SenderFeature{
						Address: tpkg.RandEd25519Address(),
					},
				},
			},
			target:    &iotago.BasicOutput{},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deSeriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestOutputsSyntacticalDepositAmount(t *testing.T) {
	protoParams := tpkg.IOTAMainnetV3TestProtocolParameters

	var minAmount iotago.BaseToken = 14100

	tests := []struct {
		name        string
		protoParams iotago.ProtocolParameters
		outputs     iotago.Outputs[iotago.Output]
		wantErr     error
	}{
		{
			name:        "ok",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:           protoParams.TokenSupply(),
					UnlockConditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()}},
					Mana:             500,
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - storage deposit covered",
			protoParams: protoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount:           minAmount, // min amount
					UnlockConditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()}},
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - storage deposit return",
			protoParams: protoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 100000,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
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
			protoParams: protoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 100000,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
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
			protoParams: protoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							// off by one from the deposit
							Amount: OneIOTA + 1,
						},
					},
					Mana: 500,
				},
			},
			wantErr: iotago.ErrStorageDepositExceedsTargetOutputAmount,
		},
		{
			name:        "fail - storage deposit not covered",
			protoParams: protoParams,
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: minAmount - 1,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrStorageDepositNotCovered,
		},
		{
			name:        "fail - zero deposit",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: 0,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrAmountMustBeGreaterThanZero,
		},
		{
			name:        "fail - more than total supply on single output",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: protoParams.TokenSupply() + 1,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrOutputsSumExceedsTotalSupply,
		},
		{
			name:        "fail - sum more than total supply over multiple outputs",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.BasicOutput{
					Amount: protoParams.TokenSupply() - 1,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.BasicOutput{
					Amount: protoParams.TokenSupply() - 1,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrOutputsSumExceedsTotalSupply,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.OutputsSyntacticalDepositAmount(tt.protoParams, iotago.NewStorageScoreStructure(tt.protoParams.StorageScoreParameters()))
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			fmt.Println(tt.name)
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
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          1337,
						},
					},
				},
				&iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.TimelockUnlockCondition{
							Slot: 1337,
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
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          0,
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
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.TimelockUnlockCondition{
							Slot: 0,
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
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				},
				&iotago.BasicOutput{
					Amount: 1,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
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
					UnlockConditions: iotago.BasicOutputUnlockConditions{
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
					Amount:         OneIOTA,
					AccountID:      iotago.AccountID{},
					FoundryCounter: 0,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non empty state",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - foundry counter non zero on empty account ID",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      iotago.AccountID{},
					FoundryCounter: 1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: iotago.ErrAccountOutputNonEmptyState,
		},
		{
			name: "fail - account's unlock condition contains its own account address",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AccountOutput {
					accountID := iotago.AccountID(tpkg.Rand32ByteArray())

					return &iotago.AccountOutput{
						Amount:         OneIOTA,
						AccountID:      accountID,
						FoundryCounter: 1337,
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: accountID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAccountOutputCyclicAddress,
		},
		{
			name: "ok - staked amount equal to amount",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{StakedAmount: OneIOTA},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - staked amount less than amount",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneIOTA + 1,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{StakedAmount: OneIOTA},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - staked amount greater than amount",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{StakedAmount: OneIOTA + 1},
					},
				},
			},
			wantErr: iotago.ErrAccountOutputAmountLessThanStakedAmount,
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
					Amount:     OneIOTA,
					AnchorID:   iotago.AnchorID{},
					StateIndex: 0,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non empty state",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AnchorOutput{
					Amount:     OneIOTA,
					AnchorID:   tpkg.Rand32ByteArray(),
					StateIndex: 10,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state index non zero on empty anchor ID",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.AnchorOutput{
					Amount:     OneIOTA,
					AnchorID:   iotago.AnchorID{},
					StateIndex: 1,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
					},
				},
			},
			wantErr: iotago.ErrAnchorOutputNonEmptyState,
		},
		{
			name: "fail - anchors's state controller address unlock condition contains its own anchor address",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AnchorOutput {
					anchorID := iotago.AnchorID(tpkg.Rand32ByteArray())

					return &iotago.AnchorOutput{
						Amount:     OneIOTA,
						AnchorID:   anchorID,
						StateIndex: 10,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: anchorID.ToAddress()},
							&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						},
					}
				}(),
			},
			wantErr: iotago.ErrAnchorOutputCyclicAddress,
		},
		{
			name: "fail - anchors's governor address unlock condition contains its own anchor address",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.AnchorOutput {
					anchorID := iotago.AnchorID(tpkg.Rand32ByteArray())

					return &iotago.AnchorOutput{
						Amount:     OneIOTA,
						AnchorID:   anchorID,
						StateIndex: 10,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
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
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
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
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
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
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
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
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
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
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
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
					Amount: OneIOTA,
					NFTID:  iotago.NFTID{},
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - NFT's address unlock condition contains its own NFT address",
			outputs: iotago.Outputs[iotago.Output]{
				func() *iotago.NFTOutput {
					nftID := iotago.NFTID(tpkg.Rand32ByteArray())

					return &iotago.NFTOutput{
						Amount: OneIOTA,
						NFTID:  nftID,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
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
					Amount:           OneIOTA,
					DelegatedAmount:  OneIOTA,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - Delegation Output contains empty validator address",
			outputs: iotago.Outputs[iotago.Output]{
				&iotago.DelegationOutput{
					Amount:           OneIOTA,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: &emptyAccountAddress,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: iotago.ErrDelegationValidatorAddressEmpty,
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
	tests := []*test{
		func() *test {
			receiverIdent := tpkg.RandEd25519Address()
			return &test{
				name: "can unlock - target is source (no unlock conditions)",
				output: &iotago.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
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
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
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
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          26,
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
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: returnIdent,
							Slot:          25,
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
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: returnIdent,
							Slot:          15,
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
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: returnIdent,
							Slot:          16,
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
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.TimelockUnlockCondition{Slot: 15},
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
					Amount: OneIOTA,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: receiverIdent},
						&iotago.TimelockUnlockCondition{Slot: 16},
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

func TestAnchorOutput_UnlockableBy(t *testing.T) {
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
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AnchorOutput{
					Amount:     OneIOTA,
					StateIndex: 1,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
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
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &iotago.AnchorOutput{
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
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
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
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

func TestOutputsSyntacticDisallowedImplicitAccountCreationAddress(t *testing.T) {
	type test struct {
		name    string
		output  iotago.Output
		wantErr error
	}

	tests := []test{
		{
			name: "fail - Account Output contains Implicit Account Creation Address",
			output: &iotago.AccountOutput{
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Anchor Output contains Implicit Account Creation Address as State Controller",
			output: &iotago.AnchorOutput{
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			wantErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Anchor Output contains Implicit Account Creation Address as Governor",
			output: &iotago.AnchorOutput{
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - NFT Output contains Implicit Account Creation Address",
			output: &iotago.NFTOutput{
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Delegation Output contains Implicit Account Creation Address",
			output: &iotago.DelegationOutput{
				Amount:           1337,
				DelegatedAmount:  1337,
				DelegationID:     tpkg.Rand32ByteArray(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       iotago.EpochIndex(32),
				EndEpoch:         iotago.EpochIndex(37),
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: iotago.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			implicitAccountCreationAddressValidatorFunc := iotago.OutputsSyntacticalImplicitAccountCreationAddress()

			err := implicitAccountCreationAddressValidatorFunc(0, tt.output)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}

}
