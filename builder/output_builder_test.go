package builder_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestBasicOutputBuilder(t *testing.T) {
	var (
		targetAddr                          = tpkg.RandEd25519Address()
		amount             iotago.BaseToken = 1337
		nativeTokenFeature                  = tpkg.RandNativeTokenFeature()
		expirationTarget                    = tpkg.RandEd25519Address()
		metadata                            = []byte("123456")
		slotTimeProvider                    = iotago.NewTimeProvider(time.Now().Unix(), 10, 10)
	)
	timelock := slotTimeProvider.SlotFromTime(time.Now().Add(5 * time.Minute))
	expiration := slotTimeProvider.SlotFromTime(time.Now().Add(10 * time.Minute))

	basicOutput, err := builder.NewBasicOutputBuilder(targetAddr, amount).
		NativeToken(nativeTokenFeature).
		Timelock(timelock).
		Expiration(expirationTarget, expiration).
		Metadata(metadata).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.BasicOutput{
		Amount: 1337,
		Conditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddr},
			&iotago.TimelockUnlockCondition{Slot: timelock},
			&iotago.ExpirationUnlockCondition{ReturnAddress: expirationTarget, Slot: expiration},
		},
		Features: iotago.BasicOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
			nativeTokenFeature,
		},
	}, basicOutput)
}

func TestAccountOutputBuilder(t *testing.T) {
	var (
		stateCtrl                    = tpkg.RandEd25519Address()
		gov                          = tpkg.RandEd25519Address()
		amount      iotago.BaseToken = 1337
		metadata                     = []byte("123456")
		immMetadata                  = []byte("654321")
		immSender                    = tpkg.RandEd25519Address()

		blockIssuerKey1    = iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		blockIssuerKey2    = iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		blockIssuerKey3    = iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		newBlockIssuerKey1 = iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		newBlockIssuerKey2 = iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
	)

	accountOutput, err := builder.NewAccountOutputBuilder(stateCtrl, gov, amount).
		Metadata(metadata).
		StateMetadata(metadata).
		Staking(amount, 1, 1000).
		BlockIssuer(iotago.NewBlockIssuerKeys(blockIssuerKey1, blockIssuerKey2, blockIssuerKey3), 100000).
		ImmutableMetadata(immMetadata).
		ImmutableSender(immSender).
		FoundriesToGenerate(5).
		Build()
	require.NoError(t, err)

	expectedBlockIssuerKeys := iotago.NewBlockIssuerKeys(blockIssuerKey1, blockIssuerKey2, blockIssuerKey3)

	expected := &iotago.AccountOutput{
		Amount:         1337,
		StateIndex:     0,
		StateMetadata:  metadata,
		FoundryCounter: 5,
		Conditions: iotago.AccountOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: gov},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
			&iotago.BlockIssuerFeature{
				BlockIssuerKeys: expectedBlockIssuerKeys,
				ExpirySlot:      100000,
			},
			&iotago.StakingFeature{
				StakedAmount: amount,
				FixedCost:    1,
				StartEpoch:   1000,
				EndEpoch:     iotago.MaxEpochIndex,
			},
		},
		ImmutableFeatures: iotago.AccountOutputImmFeatures{
			&iotago.SenderFeature{Address: immSender},
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}
	require.True(t, expected.Equal(accountOutput), "account output should be equal")

	const newAmount iotago.BaseToken = 7331
	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	expectedCpy := expected.Clone().(*iotago.AccountOutput)
	expectedCpy.Amount = newAmount
	expectedCpy.StateIndex++
	updatedOutput, err := builder.NewAccountOutputBuilderFromPrevious(accountOutput).StateTransition().
		Amount(newAmount).Builder().Build()
	require.NoError(t, err)
	require.Equal(t, expectedCpy, updatedOutput)

	updatedFeatures, err := builder.NewAccountOutputBuilderFromPrevious(accountOutput).GovernanceTransition().
		BlockIssuerTransition().
		AddKeys(newBlockIssuerKey2, newBlockIssuerKey1).
		RemoveKey(blockIssuerKey3).
		RemoveKey(blockIssuerKey1).
		ExpirySlot(1500).
		GovernanceTransition().
		StakingTransition().
		EndEpoch(2000).
		Builder().Build()
	require.NoError(t, err)

	expectedUpdatedBlockIssuerKeys := iotago.NewBlockIssuerKeys(blockIssuerKey2, newBlockIssuerKey1, newBlockIssuerKey2)

	expectedFeatures := &iotago.AccountOutput{
		Amount:         1337,
		StateIndex:     0,
		StateMetadata:  metadata,
		FoundryCounter: 5,
		Conditions: iotago.AccountOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: gov},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
			&iotago.BlockIssuerFeature{
				BlockIssuerKeys: expectedUpdatedBlockIssuerKeys,
				ExpirySlot:      1500,
			},
			&iotago.StakingFeature{
				StakedAmount: amount,
				FixedCost:    1,
				StartEpoch:   1000,
				EndEpoch:     2000,
			},
		},
		ImmutableFeatures: iotago.AccountOutputImmFeatures{
			&iotago.SenderFeature{Address: immSender},
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}
	require.True(t, expectedFeatures.Equal(updatedFeatures), "features should be equal")
}

func TestDelegationOutputBuilder(t *testing.T) {
	var (
		address                           = tpkg.RandEd25519Address()
		updatedAddress                    = tpkg.RandEd25519Address()
		amount           iotago.BaseToken = 1337
		updatedAmount    iotago.BaseToken = 127
		validatorAddress                  = tpkg.RandAccountAddress()
		delegationID                      = tpkg.RandDelegationID()
	)

	delegationOutput, err := builder.NewDelegationOutputBuilder(validatorAddress, address, amount).
		DelegatedAmount(amount).
		StartEpoch(1000).
		Build()
	require.NoError(t, err)

	expected := &iotago.DelegationOutput{
		Amount:           1337,
		DelegatedAmount:  1337,
		DelegationID:     iotago.EmptyDelegationID(),
		ValidatorAddress: validatorAddress,
		StartEpoch:       1000,
		EndEpoch:         0,
		Conditions: iotago.DelegationOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: address},
		},
	}
	require.Equal(t, expected, delegationOutput)

	updatedOutput, err := builder.NewDelegationOutputBuilderFromPrevious(delegationOutput).
		DelegationID(delegationID).
		DelegatedAmount(updatedAmount).
		Amount(updatedAmount).
		EndEpoch(1500).
		Address(updatedAddress).
		Build()
	require.NoError(t, err)

	expectedOutput := &iotago.DelegationOutput{
		Amount:           127,
		DelegatedAmount:  127,
		ValidatorAddress: validatorAddress,
		DelegationID:     delegationID,
		StartEpoch:       1000,
		EndEpoch:         1500,
		Conditions: iotago.DelegationOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: updatedAddress},
		},
	}
	require.Equal(t, expectedOutput, updatedOutput)
}

func TestFoundryOutputBuilder(t *testing.T) {
	var (
		accountAddr                  = tpkg.RandAccountAddress()
		amount      iotago.BaseToken = 1337
		tokenScheme                  = &iotago.SimpleTokenScheme{
			MintedTokens:  big.NewInt(0),
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: big.NewInt(1000),
		}
		nativeTokenFeature = tpkg.RandNativeTokenFeature()
		metadata           = []byte("123456")
		immMetadata        = []byte("654321")
	)

	foundryOutput, err := builder.NewFoundryOutputBuilder(accountAddr, tokenScheme, amount).
		NativeToken(nativeTokenFeature).
		Metadata(metadata).
		ImmutableMetadata(immMetadata).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.FoundryOutput{
		Amount:      1337,
		TokenScheme: tokenScheme,
		Conditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: accountAddr},
		},
		Features: iotago.FoundryOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
			nativeTokenFeature,
		},
		ImmutableFeatures: iotago.FoundryOutputImmFeatures{
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}, foundryOutput)
}

func TestNFTOutputBuilder(t *testing.T) {
	var (
		targetAddr                   = tpkg.RandAccountAddress()
		amount      iotago.BaseToken = 1337
		metadata                     = []byte("123456")
		immMetadata                  = []byte("654321")
	)

	nftOutput, err := builder.NewNFTOutputBuilder(targetAddr, amount).
		Metadata(metadata).
		ImmutableMetadata(immMetadata).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.NFTOutput{
		Amount: 1337,
		Conditions: iotago.NFTOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddr},
		},
		Features: iotago.NFTOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
		},
		ImmutableFeatures: iotago.NFTOutputImmFeatures{
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}, nftOutput)
}
