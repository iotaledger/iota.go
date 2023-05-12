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
		Conditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddr},
			&iotago.TimelockUnlockCondition{UnixTime: uint32(timelock)},
			&iotago.ExpirationUnlockCondition{ReturnAddress: expirationTarget, UnixTime: uint32(expiration)},
		},
		Features: iotago.BasicOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
		},
	}, basicOutput)
}

func TestAccountOutputBuilder(t *testing.T) {
	var (
		stateCtrl          = tpkg.RandEd25519Address()
		gov                = tpkg.RandEd25519Address()
		deposit     uint64 = 1337
		nt                 = tpkg.RandNativeToken()
		metadata           = []byte("123456")
		immMetadata        = []byte("654321")
		immSender          = tpkg.RandEd25519Address()
	)

	accountOutput, err := builder.NewAccountOutputBuilder(stateCtrl, gov, deposit).
		NativeToken(nt).
		Metadata(metadata).
		StateMetadata(metadata).
		ImmutableMetadata(immMetadata).
		ImmutableSender(immSender).
		FoundriesToGenerate(5).
		Build()
	require.NoError(t, err)

	expected := &iotago.AccountOutput{
		Amount:         1337,
		NativeTokens:   iotago.NativeTokens{nt},
		StateIndex:     1,
		StateMetadata:  metadata,
		FoundryCounter: 5,
		Conditions: iotago.AccountOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: gov},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
		},
		ImmutableFeatures: iotago.AccountOutputImmFeatures{
			&iotago.SenderFeature{Address: immSender},
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}
	require.Equal(t, expected, accountOutput)

	const newDeposit uint64 = 7331
	expectedCpy := expected.Clone().(*iotago.AccountOutput)
	expectedCpy.Amount = newDeposit
	expectedCpy.StateIndex++
	updatedOutput, err := builder.NewAccountOutputBuilderFromPrevious(accountOutput).StateTransition().
		Deposit(newDeposit).Builder().Build()
	require.NoError(t, err)
	require.Equal(t, expectedCpy, updatedOutput)
}

func TestFoundryOutputBuilder(t *testing.T) {
	var (
		accountAddr        = tpkg.RandAccountAddress()
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

	foundryOutput, err := builder.NewFoundryOutputBuilder(accountAddr, tokenScheme, deposit).
		NativeToken(nt).
		Metadata(metadata).
		ImmutableMetadata(immMetadata).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.FoundryOutput{
		Amount:       1337,
		TokenScheme:  tokenScheme,
		NativeTokens: iotago.NativeTokens{nt},
		Conditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: accountAddr},
		},
		Features: iotago.FoundryOutputFeatures{
			&iotago.MetadataFeature{Data: metadata},
		},
		ImmutableFeatures: iotago.FoundryOutputImmFeatures{
			&iotago.MetadataFeature{Data: immMetadata},
		},
	}, foundryOutput)
}

func TestNFTOutputBuilder(t *testing.T) {
	var (
		targetAddr         = tpkg.RandAccountAddress()
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
