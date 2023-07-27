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
	ntIn := &iotago.NativeToken{
		ID:     tpkg.Rand38ByteArray(),
		Amount: new(big.Int).SetUint64(1000),
	}

	ntBytes, err := tpkg.TestAPI.Encode(ntIn, serix.WithValidation())
	require.NoError(t, err)

	ntOut := &iotago.NativeToken{}
	_, err = tpkg.TestAPI.Decode(ntBytes, ntOut, serix.WithValidation())
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
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.AccountOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					AccountID:    iotago.AccountID{},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.FoundryOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  tpkg.RandUint256(),
						MeltedTokens:  tpkg.RandUint256(),
						MaximumSupply: tpkg.RandUint256(),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
				&iotago.NFTOutput{
					Amount:       1,
					NativeTokens: tpkg.RandSortNativeTokens(5),
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name:    "not sorted",
			wantErr: true,
			sources: []iotago.Output{
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: notSortedNativeTokens(),
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.AccountOutput{
					Amount:       1,
					NativeTokens: notSortedNativeTokens(),
					AccountID:    iotago.AccountID{},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.FoundryOutput{
					Amount:       1,
					NativeTokens: notSortedNativeTokens(),
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  tpkg.RandUint256(),
						MeltedTokens:  tpkg.RandUint256(),
						MaximumSupply: tpkg.RandUint256(),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
				&iotago.NFTOutput{
					Amount:       1,
					NativeTokens: notSortedNativeTokens(),
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name:    "duped",
			wantErr: true,
			sources: []iotago.Output{
				&iotago.BasicOutput{
					Amount:       1,
					NativeTokens: dupedNativeTokens(),
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.AccountOutput{
					Amount:       1,
					NativeTokens: dupedNativeTokens(),
					AccountID:    iotago.AccountID{},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&iotago.FoundryOutput{
					Amount:       1,
					NativeTokens: dupedNativeTokens(),
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  tpkg.RandUint256(),
						MeltedTokens:  tpkg.RandUint256(),
						MaximumSupply: tpkg.RandUint256(),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
				&iotago.NFTOutput{
					Amount:       1,
					NativeTokens: dupedNativeTokens(),
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, source := range test.sources {
				if _, err := tpkg.TestAPI.Encode(source, serix.WithValidation()); (err != nil) != test.wantErr {
					t.Errorf("error = %v, wantErr %v", err, test.wantErr)
				}
			}
		})
	}
}
