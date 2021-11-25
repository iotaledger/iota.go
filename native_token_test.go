package iotago_test

import (
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestNativeTokenDeSerialization(t *testing.T) {
	ntIn := iotago.NativeToken{
		ID:     tpkg.Rand38ByteArray(),
		Amount: new(big.Int).SetUint64(1000),
	}

	ntBytes, err := ntIn.Serialize(serializer.DeSeriModeNoValidation, nil)
	require.NoError(t, err)

	var ntOut iotago.NativeToken
	_, err = ntOut.Deserialize(ntBytes, serializer.DeSeriModeNoValidation, nil)
	require.NoError(t, err)

	require.EqualValues(t, ntIn, ntOut)
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
				if _, err := source.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas); (err != nil) != test.wantErr {
					t.Errorf("error = %v, wantErr %v", err, test.wantErr)
				}
			}
		})
	}
}
