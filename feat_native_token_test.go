//nolint:dupl,scopelint
package iotago_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestNativeTokenDeSerialization(t *testing.T) {
	ntIn := &iotago.NativeTokenFeature{
		ID:     tpkg.Rand38ByteArray(),
		Amount: new(big.Int).SetUint64(1000),
	}

	ntBytes, err := tpkg.ZeroCostTestAPI.Encode(ntIn, serix.WithValidation())
	require.NoError(t, err)

	ntOut := &iotago.NativeTokenFeature{}
	_, err = tpkg.ZeroCostTestAPI.Decode(ntBytes, ntOut, serix.WithValidation())
	require.NoError(t, err)

	require.EqualValues(t, ntIn, ntOut)
}

func TestNativeToken_SyntacticalValidation(t *testing.T) {
	nativeTokenFeature := tpkg.RandNativeTokenFeature()
	accountAddress, err := nativeTokenFeature.ID.AccountAddress()
	require.NoError(t, err)

	type test struct {
		name               string
		nativeTokenFeature *iotago.NativeTokenFeature
		wantErr            error
	}

	tests := []*test{
		{
			name:               "ok - NativeTokenFeature token ID == FoundryID",
			nativeTokenFeature: nativeTokenFeature,
			wantErr:            nil,
		},
		{
			name:               "fail - NativeTokenFeature token ID != FoundryID",
			nativeTokenFeature: tpkg.RandNativeTokenFeature(),
			wantErr:            iotago.ErrFoundryIDNativeTokenIDMismatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			foundryIn := &iotago.FoundryOutput{
				Amount:       1000,
				SerialNumber: nativeTokenFeature.ID.FoundrySerialNumber(),
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  big.NewInt(100),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: big.NewInt(100),
				},
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{
						Address: accountAddress,
					},
				},
				Features: iotago.FoundryOutputFeatures{
					test.nativeTokenFeature,
				},
				ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
			}

			foundryBytes, err := tpkg.ZeroCostTestAPI.Encode(foundryIn, serix.WithValidation())
			if err == nil {
				err = iotago.OutputsSyntacticalFoundry()(0, foundryIn)
			}
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)

			foundryOut := &iotago.FoundryOutput{}
			_, err = tpkg.ZeroCostTestAPI.Decode(foundryBytes, foundryOut, serix.WithValidation())
			require.NoError(t, err)

			require.True(t, foundryIn.Equal(foundryOut))
		})
	}
}
