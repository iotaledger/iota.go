package iotago_test

import (
	"reflect"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3"
)

func TestOutputsNativeTokenSet(t *testing.T) {
	notSortedNativeTokens := func() iotago.NativeTokens {
		nativeTokens := tpkg.RandSortNativeTokens(5)
		nativeTokens[0], nativeTokens[1] = nativeTokens[1], nativeTokens[0]
		return nativeTokens
	}

	dupedNativeTokens := func() iotago.NativeTokens {
		nativeTokens := tpkg.RandSortNativeTokens(2)
		nativeTokens[1] = nativeTokens[0]
		return nativeTokens
	}

	tests := []struct {
		name    string
		wantErr bool
		sources []iotago.Output
	}{
		{
			name:    "ok",
			wantErr: false,
			sources: []iotago.Output{
				&iotago.ExtendedOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					Address:      tpkg.RandEd25519Address(),
				},
				&iotago.AliasOutput{
					Amount:               1,
					NativeTokens:         tpkg.RandSortNativeTokens(5),
					AliasID:              iotago.AliasID{},
					StateController:      tpkg.RandEd25519Address(),
					GovernanceController: tpkg.RandEd25519Address(),
				},
				&iotago.FoundryOutput{
					Amount:            1,
					NativeTokens:      tpkg.RandSortNativeTokens(5),
					Address:           tpkg.RandAliasAddress(),
					CirculatingSupply: tpkg.RandUint256(),
					MaximumSupply:     tpkg.RandUint256(),
					TokenScheme:       &iotago.SimpleTokenScheme{},
				},
				&iotago.NFTOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					Address:      tpkg.RandEd25519Address(),
				},
			},
		},
		{
			name:    "not sorted",
			wantErr: true,
			sources: []iotago.Output{
				&iotago.ExtendedOutput{
					Amount:       1,
					NativeTokens: notSortedNativeTokens(),
					Address:      tpkg.RandEd25519Address(),
				},
				&iotago.AliasOutput{
					Amount:               1,
					NativeTokens:         notSortedNativeTokens(),
					AliasID:              iotago.AliasID{},
					StateController:      tpkg.RandEd25519Address(),
					GovernanceController: tpkg.RandEd25519Address(),
				},
				&iotago.FoundryOutput{
					Amount:            1,
					NativeTokens:      notSortedNativeTokens(),
					Address:           tpkg.RandEd25519Address(),
					CirculatingSupply: tpkg.RandUint256(),
					MaximumSupply:     tpkg.RandUint256(),
					TokenScheme:       &iotago.SimpleTokenScheme{},
				},
				&iotago.NFTOutput{
					Amount:       1,
					NativeTokens: notSortedNativeTokens(),
					Address:      tpkg.RandEd25519Address(),
				},
			},
		},
		{
			name:    "duped",
			wantErr: true,
			sources: []iotago.Output{
				&iotago.ExtendedOutput{
					Amount:       1,
					NativeTokens: dupedNativeTokens(),
					Address:      tpkg.RandEd25519Address(),
				},
				&iotago.AliasOutput{
					Amount:               1,
					NativeTokens:         dupedNativeTokens(),
					AliasID:              iotago.AliasID{},
					StateController:      tpkg.RandEd25519Address(),
					GovernanceController: tpkg.RandEd25519Address(),
				},
				&iotago.FoundryOutput{
					Amount:            1,
					NativeTokens:      dupedNativeTokens(),
					Address:           tpkg.RandEd25519Address(),
					CirculatingSupply: tpkg.RandUint256(),
					MaximumSupply:     tpkg.RandUint256(),
					TokenScheme:       &iotago.SimpleTokenScheme{},
				},
				&iotago.NFTOutput{
					Amount:       1,
					NativeTokens: dupedNativeTokens(),
					Address:      tpkg.RandEd25519Address(),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, source := range test.sources {
				if _, err := source.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas); (err != nil) != test.wantErr {
					t.Errorf("error = %v, wantErr %v", err, test.wantErr)
				}
			}
		})
	}
}

func TestOutputsSyntacticalDepositAmount(t *testing.T) {

	nonZeroCostParas := &iotago.DeSerializationParameters{
		RentStructure: &iotago.RentStructure{
			VByteCost:    1,
			VBFactorData: iotago.VByteCostFactorData,
			VBFactorKey:  iotago.VByteCostFactorKey,
		},
		MinDustDeposit: 580,
	}

	tests := []struct {
		name        string
		deSeriParas *iotago.DeSerializationParameters
		outputs     iotago.Outputs
		wantErr     error
	}{
		{
			name:        "ok",
			deSeriParas: DefZeroRentParas,
			outputs: iotago.Outputs{
				&iotago.SimpleOutput{Amount: iotago.TokenSupply, Address: tpkg.RandEd25519Address()},
			},
			wantErr: nil,
		},
		{
			name:        "ok - state rent covered",
			deSeriParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.SimpleOutput{Amount: 580, Address: tpkg.RandAliasAddress()},
			},
			wantErr: nil,
		},
		{
			name:        "ok - dust deposit return",
			deSeriParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandAliasAddress(),
					Amount:       OneMi * 2,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{
							Amount: 592,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:        "fail - dust deposit return more than state rent",
			deSeriParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandAliasAddress(),
					Amount:       OneMi * 2,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{
							Amount: 593, // off by 1
						},
					},
				},
			},
			wantErr: iotago.ErrOutputReturnBlockIsMoreThanVBRent,
		},
		{
			name:        "fail - dust deposit return less than min dust deposit",
			deSeriParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandAliasAddress(),
					Amount:       OneMi * 2,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{
							Amount: 579, // off by 1
						},
					},
				},
			},
			wantErr: iotago.ErrOutputReturnBlockIsLessThanMinDust,
		},
		{
			name:        "fail - state rent not covered",
			deSeriParas: nonZeroCostParas,
			outputs: iotago.Outputs{
				&iotago.SimpleOutput{Amount: 100, Address: tpkg.RandAliasAddress()},
			},
			wantErr: iotago.ErrVByteRentNotCovered,
		},
		{
			name:        "fail - zero deposit",
			deSeriParas: DefZeroRentParas,
			outputs: iotago.Outputs{
				&iotago.SimpleOutput{Amount: 0, Address: tpkg.RandEd25519Address()},
			},
			wantErr: iotago.ErrDepositAmountMustBeGreaterThanZero,
		},
		{
			name:        "fail - more than total supply on single output",
			deSeriParas: DefZeroRentParas,
			outputs: iotago.Outputs{
				&iotago.SimpleOutput{Amount: iotago.TokenSupply + 1, Address: tpkg.RandEd25519Address()},
			},
			wantErr: iotago.ErrOutputDepositsMoreThanTotalSupply,
		},
		{
			name:        "fail - sum more than total supply over multiple outputs",
			deSeriParas: DefZeroRentParas,
			outputs: iotago.Outputs{
				&iotago.SimpleOutput{Amount: iotago.TokenSupply - 1, Address: tpkg.RandEd25519Address()},
				&iotago.SimpleOutput{Amount: iotago.TokenSupply - 1, Address: tpkg.RandEd25519Address()},
			},
			wantErr: iotago.ErrOutputsSumExceedsTotalSupply,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.OutputsSyntacticalDepositAmount(tt.deSeriParas.MinDustDeposit, tt.deSeriParas.RentStructure)
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
				&iotago.ExtendedOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					Address:      tpkg.RandEd25519Address(),
				},
				&iotago.ExtendedOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(10),
					Address:      tpkg.RandEd25519Address(),
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - sum more than max native tokens count",
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(200),
					Address:      tpkg.RandEd25519Address(),
				},
				&iotago.ExtendedOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(200),
					Address:      tpkg.RandEd25519Address(),
				},
			},
			wantErr: iotago.ErrOutputsExceedMaxNativeTokensCount,
		},
	}
	valFunc := iotago.OutputsSyntacticalNativeTokensCount()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestOutputsSyntacticalSenderFeatureBlockRequirement(t *testing.T) {
	tests := []struct {
		name string
		want iotago.OutputsSyntacticalValidationFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := iotago.OutputsSyntacticalSenderFeatureBlockRequirement(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OutputsSyntacticalSenderFeatureBlockRequirement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputsSyntacticalAlias(t *testing.T) {
	type args struct {
		txID *iotago.TransactionID
	}
	tests := []struct {
		name string
		args args
		want iotago.OutputsSyntacticalValidationFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := iotago.OutputsSyntacticalAlias(tt.args.txID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OutputsSyntacticalAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputsSyntacticalFoundry(t *testing.T) {
	tests := []struct {
		name string
		want iotago.OutputsSyntacticalValidationFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := iotago.OutputsSyntacticalFoundry(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OutputsSyntacticalFoundry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputsSyntacticalNFT(t *testing.T) {
	type args struct {
		txID *iotago.TransactionID
	}
	tests := []struct {
		name string
		args args
		want iotago.OutputsSyntacticalValidationFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := iotago.OutputsSyntacticalNFT(tt.args.txID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OutputsSyntacticalNFT() = %v, want %v", got, tt.want)
			}
		})
	}
}
