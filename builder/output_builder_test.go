package builder_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestBasicOutputBuilder(t *testing.T) {
	var (
		targetAddr              = tpkg.RandEd25519Address()
		deposit          uint64 = 1337
		nt                      = tpkg.RandNativeToken()
		timelock                = time.Now().Add(5 * time.Minute).Unix()
		expiration              = time.Now().Add(10 * time.Minute).Unix()
		expirationTarget        = tpkg.RandEd25519Address()
		metadata                = []byte("123456")
	)

	basicOutput, err := builder.NewBasicOutputBuilder(targetAddr, deposit).
		NativeToken(nt).
		Timelock(timelock).
		Expiration(expirationTarget, expiration).
		Metadata(metadata).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.BasicOutput{
		Amount:       1337,
		NativeTokens: iotago.NativeTokens{nt},
		Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
			&iotago.AddressUnlockCondition{Address: targetAddr},
			&iotago.TimelockUnlockCondition{UnixTime: uint32(timelock)},
			&iotago.ExpirationUnlockCondition{ReturnAddress: expirationTarget, UnixTime: uint32(expiration)},
		},
		Features: iotago.Features[iotago.BasicOutputFeature]{
			&iotago.MetadataFeature{Data: metadata},
		},
	}, basicOutput)
}

func TestAliasOutputBuilder(t *testing.T) {
	var (
		stateCtrl          = tpkg.RandEd25519Address()
		gov                = tpkg.RandEd25519Address()
		deposit     uint64 = 1337
		nt                 = tpkg.RandNativeToken()
		metadata           = []byte("123456")
		immMetadata        = []byte("654321")
		immSender          = tpkg.RandEd25519Address()
	)

	aliasOutput, err := builder.NewAliasOutputBuilder(stateCtrl, gov, deposit).
		NativeToken(nt).
		Metadata(metadata).
		StateMetadata(metadata).
		ImmutableMetadata(immMetadata).
		ImmutableSender(immSender).
		FoundriesToGenerate(5).
		Build()
	require.NoError(t, err)

	expected := &iotago.AliasOutput{
		Amount:         1337,
		NativeTokens:   iotago.NativeTokens{nt},
		StateIndex:     1,
		StateMetadata:  metadata,
		FoundryCounter: 5,
		Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: gov},
		},
		Features: iotago.Features[iotago.AliasFeature]{
			&iotago.MetadataFeature{Data: metadata},
		},
		ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{
			&iotago.SenderFeature{Address: immSender},
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}
	require.Equal(t, expected, aliasOutput)

	const newDeposit uint64 = 7331
	expectedCpy := expected.Clone().(*iotago.AliasOutput)
	expectedCpy.Amount = newDeposit
	expectedCpy.StateIndex++
	updatedOutput, err := builder.NewAliasOutputBuilderFromPrevious(aliasOutput).StateTransition().
		Deposit(newDeposit).Builder().Build()
	require.NoError(t, err)
	require.Equal(t, expectedCpy, updatedOutput)
}

func TestFoundryOutputBuilder(t *testing.T) {
	var (
		aliasAddr          = tpkg.RandAliasAddress()
		deposit     uint64 = 1337
		tokenScheme        = &iotago.SimpleTokenScheme{
			MintedTokens:  big.NewInt(0),
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: big.NewInt(1000),
		}
		nt          = tpkg.RandNativeToken()
		metadata    = []byte("123456")
		immMetadata = []byte("654321")
	)

	foundryOutput, err := builder.NewFoundryOutputBuilder(aliasAddr, tokenScheme, deposit).
		NativeToken(nt).
		Metadata(metadata).
		ImmutableMetadata(immMetadata).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.FoundryOutput{
		Amount:       1337,
		TokenScheme:  tokenScheme,
		NativeTokens: iotago.NativeTokens{nt},
		Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
			&iotago.ImmutableAliasUnlockCondition{Address: aliasAddr},
		},
		Features: iotago.Features[iotago.FoundryFeature]{
			&iotago.MetadataFeature{Data: metadata},
		},
		ImmutableFeatures: iotago.Features[iotago.FoundryImmFeature]{
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}, foundryOutput)
}

func TestNFTOutputBuilder(t *testing.T) {
	var (
		targetAddr         = tpkg.RandAliasAddress()
		deposit     uint64 = 1337
		nt                 = tpkg.RandNativeToken()
		metadata           = []byte("123456")
		immMetadata        = []byte("654321")
	)

	nftOutput, err := builder.NewNFTOutputBuilder(targetAddr, deposit).
		NativeToken(nt).
		Metadata(metadata).
		ImmutableMetadata(immMetadata).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.NFTOutput{
		Amount:       1337,
		NativeTokens: iotago.NativeTokens{nt},
		Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
			&iotago.AddressUnlockCondition{Address: targetAddr},
		},
		Features: iotago.Features[iotago.NFTFeature]{
			&iotago.MetadataFeature{Data: metadata},
		},
		ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}, nftOutput)
}
