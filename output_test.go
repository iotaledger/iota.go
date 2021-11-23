package iotago_test

import (
	"math/big"
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3"
)

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
		name    string
		outputs iotago.Outputs
		wantErr error
	}{
		{
			name: "ok - dust deposit return",
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandEd25519Address(),
					Amount:       100,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{Amount: OneMi},
						&iotago.SenderFeatureBlock{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - dust deposit return",
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandEd25519Address(),
					Amount:       100,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{Amount: OneMi},
					},
				},
			},
			wantErr: iotago.ErrOutputRequiresSenderFeatureBlock,
		},
		{
			name: "ok - expiration milestone index",
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandEd25519Address(),
					Amount:       100,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 1337},
						&iotago.SenderFeatureBlock{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - expiration milestone index",
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandEd25519Address(),
					Amount:       100,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 1337},
					},
				},
			},
			wantErr: iotago.ErrOutputRequiresSenderFeatureBlock,
		},
		{
			name: "ok - expiration unix",
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandEd25519Address(),
					Amount:       100,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.ExpirationUnixFeatureBlock{UnixTime: 1337},
						&iotago.SenderFeatureBlock{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - expiration unix",
			outputs: iotago.Outputs{
				&iotago.ExtendedOutput{
					Address:      tpkg.RandEd25519Address(),
					Amount:       100,
					NativeTokens: nil,
					Blocks: iotago.FeatureBlocks{
						&iotago.ExpirationUnixFeatureBlock{UnixTime: 1337},
					},
				},
			},
			wantErr: iotago.ErrOutputRequiresSenderFeatureBlock,
		},
	}
	valFunc := iotago.OutputsSyntacticalSenderFeatureBlockRequirement()
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
					Amount:               OneMi,
					AliasID:              iotago.AliasID{},
					StateController:      tpkg.RandAliasAddress(),
					GovernanceController: tpkg.RandAliasAddress(),
					StateIndex:           0,
					FoundryCounter:       0,
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non empty state",
			outputs: iotago.Outputs{
				&iotago.AliasOutput{
					Amount:               OneMi,
					AliasID:              tpkg.Rand20ByteArray(),
					StateController:      tpkg.RandAliasAddress(),
					GovernanceController: tpkg.RandAliasAddress(),
					StateIndex:           10,
					FoundryCounter:       1337,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state index non zero on empty alias ID",
			outputs: iotago.Outputs{
				&iotago.AliasOutput{
					Amount:               OneMi,
					AliasID:              iotago.AliasID{},
					StateController:      tpkg.RandAliasAddress(),
					GovernanceController: tpkg.RandAliasAddress(),
					StateIndex:           1,
					FoundryCounter:       0,
				},
			},
			wantErr: iotago.ErrAliasOutputNonEmptyState,
		},
		{
			name: "fail - foundry counter non zero on empty alias ID",
			outputs: iotago.Outputs{
				&iotago.AliasOutput{
					Amount:               OneMi,
					AliasID:              iotago.AliasID{},
					StateController:      tpkg.RandAliasAddress(),
					GovernanceController: tpkg.RandAliasAddress(),
					StateIndex:           0,
					FoundryCounter:       1,
				},
			},
			wantErr: iotago.ErrAliasOutputNonEmptyState,
		},
		{
			name: "fail - cyclic state controller",
			outputs: iotago.Outputs{
				func() *iotago.AliasOutput {
					aliasID := iotago.AliasID(tpkg.Rand20ByteArray())
					return &iotago.AliasOutput{
						Amount:               OneMi,
						AliasID:              aliasID,
						StateController:      aliasID.ToAddress(),
						GovernanceController: tpkg.RandAliasAddress(),
						StateIndex:           10,
						FoundryCounter:       1337,
					}
				}(),
			},
			wantErr: iotago.ErrAliasOutputCyclicAddress,
		},
		{
			name: "fail - cyclic governance controller",
			outputs: iotago.Outputs{
				func() *iotago.AliasOutput {
					aliasID := iotago.AliasID(tpkg.Rand20ByteArray())
					return &iotago.AliasOutput{
						Amount:               OneMi,
						AliasID:              aliasID,
						StateController:      tpkg.RandAliasAddress(),
						GovernanceController: aliasID.ToAddress(),
						StateIndex:           10,
						FoundryCounter:       1337,
					}
				}(),
			},
			wantErr: iotago.ErrAliasOutputCyclicAddress,
		},
	}
	randTxID := tpkg.Rand32ByteArray()
	valFunc := iotago.OutputsSyntacticalAlias(&randTxID)
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
					Address:           tpkg.RandAliasAddress(),
					Amount:            1337,
					NativeTokens:      nil,
					SerialNumber:      5,
					TokenTag:          tpkg.Rand12ByteArray(),
					CirculatingSupply: new(big.Int).SetUint64(5),
					MaximumSupply:     new(big.Int).SetUint64(10),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - circ and max supply same",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Address:           tpkg.RandAliasAddress(),
					Amount:            1337,
					NativeTokens:      nil,
					SerialNumber:      5,
					TokenTag:          tpkg.Rand12ByteArray(),
					CirculatingSupply: new(big.Int).SetUint64(10),
					MaximumSupply:     new(big.Int).SetUint64(10),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid maximum supply",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Address:           tpkg.RandAliasAddress(),
					Amount:            1337,
					NativeTokens:      nil,
					SerialNumber:      5,
					TokenTag:          tpkg.Rand12ByteArray(),
					CirculatingSupply: new(big.Int).SetUint64(5),
					MaximumSupply:     new(big.Int).SetUint64(0),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
			},
			wantErr: iotago.ErrFoundryOutputInvalidMaximumSupply,
		},
		{
			name: "fail - invalid circulating supply",
			outputs: iotago.Outputs{
				&iotago.FoundryOutput{
					Address:           tpkg.RandAliasAddress(),
					Amount:            1337,
					NativeTokens:      nil,
					SerialNumber:      5,
					TokenTag:          tpkg.Rand12ByteArray(),
					CirculatingSupply: new(big.Int).SetUint64(5),
					MaximumSupply:     new(big.Int).SetUint64(4),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
			},
			wantErr: iotago.ErrFoundryOutputInvalidCirculatingSupply,
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
					Address: tpkg.RandEd25519Address(),
					Amount:  OneMi,
					NFTID:   iotago.NFTID{},
				},
			},
		},
		{
			name: "fail - cyclic",
			outputs: iotago.Outputs{
				func() *iotago.NFTOutput {
					nftID := iotago.NFTID(tpkg.Rand20ByteArray())
					return &iotago.NFTOutput{
						Address: nftID.ToAddress(),
						Amount:  OneMi,
						NFTID:   nftID,
					}
				}(),
			},
			wantErr: iotago.ErrNFTOutputCyclicAddress,
		},
	}
	randTxID := tpkg.Rand32ByteArray()
	valFunc := iotago.OutputsSyntacticalNFT(&randTxID)
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
