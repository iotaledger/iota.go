package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
)

var (
	zeroRentStruct = &iotago.RentStructure{
		VByteCost:    0,
		VBFactorData: 0,
		VBFactorKey:  0,
	}
)

func TestOutputSelector(t *testing.T) {
	_, err := iotago.OutputSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownOutputType))
}

func TestOutputsPredicateFuncs(t *testing.T) {
	type args struct {
		outputs iotago.Outputs
		funcs   []iotago.OutputsSyntacticalValidationFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"deposit amount - ok",
			args{outputs: iotago.Outputs{
				&iotago.SimpleOutput{
					Amount:  iotago.TokenSupply,
					Address: tpkg.RandEd25519Address(),
				},
			}, funcs: []iotago.OutputsSyntacticalValidationFunc{iotago.OutputsSyntacticalDepositAmount(0, zeroRentStruct)}}, false,
		},
		{
			"deposit amount - more than total supply",
			args{outputs: iotago.Outputs{
				&iotago.SimpleOutput{
					Amount:  iotago.TokenSupply + 1,
					Address: tpkg.RandEd25519Address(),
				},
			}, funcs: []iotago.OutputsSyntacticalValidationFunc{iotago.OutputsSyntacticalDepositAmount(0, zeroRentStruct)}}, true,
		},
		{
			"deposit amount- sum more than total supply",
			args{outputs: iotago.Outputs{
				&iotago.SimpleOutput{
					Amount:  iotago.TokenSupply - 1,
					Address: tpkg.RandEd25519Address(),
				},
				&iotago.SimpleOutput{
					Amount:  iotago.TokenSupply - 1,
					Address: tpkg.RandEd25519Address(),
				},
			}, funcs: []iotago.OutputsSyntacticalValidationFunc{iotago.OutputsSyntacticalDepositAmount(0, zeroRentStruct)}}, true,
		},
		{
			"native tokens count - ok",
			args{outputs: iotago.Outputs{
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
			}, funcs: []iotago.OutputsSyntacticalValidationFunc{iotago.OutputsSyntacticalNativeTokensCount()}}, false,
		},
		{
			"native tokens count - sum more than max native tokens count",
			args{outputs: iotago.Outputs{
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
			}, funcs: []iotago.OutputsSyntacticalValidationFunc{iotago.OutputsSyntacticalNativeTokensCount()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iotago.ValidateOutputs(tt.args.outputs, tt.args.funcs...); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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
				if _, err := source.Serialize(serializer.DeSeriModePerformValidation, DefDeSeriParas); (err != nil) != test.wantErr {
					t.Errorf("error = %v, wantErr %v", err, test.wantErr)
				}
			}
		})
	}
}
