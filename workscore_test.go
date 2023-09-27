package iotago_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTransactionEssenceWorkScore(t *testing.T) {
	keyPair := hiveEd25519.GenerateKeyPair()
	keyPair2 := hiveEd25519.GenerateKeyPair()
	// Derive a dummy account from addr.
	addr := iotago.Ed25519AddressFromPubKey(keyPair.PublicKey[:])

	output1 := &iotago.BasicOutput{
		Amount:       100000,
		NativeTokens: tpkg.RandSortNativeTokens(2),
		Conditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{
				Address: addr,
			},
		},
	}
	output2 := &iotago.AccountOutput{
		Amount:       1_000_000,
		NativeTokens: tpkg.RandSortNativeTokens(3),
		Conditions: iotago.AccountOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{
				Address: addr,
			},
			&iotago.GovernorAddressUnlockCondition{
				Address: addr,
			},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.BlockIssuerFeature{
				ExpirySlot: 300,
				BlockIssuerKeys: iotago.BlockIssuerKeys{
					iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(keyPair.PublicKey),
					iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(keyPair2.PublicKey),
				},
			},
			&iotago.StakingFeature{
				StakedAmount: 500_00,
				FixedCost:    500,
			},
		},
	}

	api := iotago.V3API(
		iotago.NewV3ProtocolParameters(
			iotago.WithNetworkOptions("TestJungle", "tgl"),
			iotago.WithSupplyOptions(tpkg.TestTokenSupply, 0, 0, 0, 0, 0, 0),
			iotago.WithWorkScoreOptions(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		),
	)

	tx, err := builder.NewTransactionBuilder(api).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output1,
		}).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output1,
		}).
		AddOutput(output1).
		AddOutput(output2).
		AddContextInput(&iotago.CommitmentInput{CommitmentID: iotago.NewSlotIdentifier(85, tpkg.Rand32ByteArray())}).
		AddContextInput(&iotago.BlockIssuanceCreditInput{AccountID: tpkg.RandAccountID()}).
		AddContextInput(&iotago.RewardInput{Index: 0}).
		IncreaseAllotment(tpkg.RandAccountID(), tpkg.RandMana(10000)+1).
		IncreaseAllotment(tpkg.RandAccountID(), tpkg.RandMana(10000)+1).
		Build(iotago.NewInMemoryAddressSigner(iotago.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])}))
	require.NoError(t, err)

	workScoreStructure := api.ProtocolParameters().WorkScoreStructure()

	workScore, err := tx.WorkScore(workScoreStructure)
	require.NoError(t, err)

	// Calculate work score as defined in TIP-45 for verification.
	expectedWorkScore := workScoreStructure.DataByte*iotago.WorkScore(tx.Size()) +
		workScoreStructure.Input*2 +
		workScoreStructure.ContextInput*3 +
		// Accounts for one Signature unlock.
		workScoreStructure.SignatureEd25519 +
		workScoreStructure.BlockIssuer +
		workScoreStructure.Staking +
		workScoreStructure.NativeToken*5 +
		workScoreStructure.Allotment*2

	require.Equal(t, expectedWorkScore, workScore, "work score expected: %d, actual: %d", expectedWorkScore, workScore)
}
