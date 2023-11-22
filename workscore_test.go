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
		Amount: 100000,
		UnlockConditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{
				Address: addr,
			},
		},
		Features: iotago.BasicOutputFeatures{
			tpkg.RandNativeTokenFeature(),
		},
	}
	output2 := &iotago.AccountOutput{
		Amount: 1_000_000,
		UnlockConditions: iotago.AccountOutputUnlockConditions{
			&iotago.AddressUnlockCondition{addr},
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
			tpkg.RandNativeTokenFeature(),
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
		AddCommitmentInput(&iotago.CommitmentInput{CommitmentID: iotago.NewCommitmentID(85, tpkg.Rand32ByteArray())}).
		AddBlockIssuanceCreditInput(&iotago.BlockIssuanceCreditInput{AccountID: tpkg.RandAccountID()}).
		AddRewardInput(&iotago.RewardInput{Index: 0}, 0).
		IncreaseAllotment(tpkg.RandAccountID(), tpkg.RandMana(10000)+1).
		IncreaseAllotment(tpkg.RandAccountID(), tpkg.RandMana(10000)+1).
		Build(iotago.NewInMemoryAddressSigner(iotago.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])}))
	require.NoError(t, err)

	workScoreParameters := api.ProtocolParameters().WorkScoreParameters()

	workScore, err := tx.WorkScore(workScoreParameters)
	require.NoError(t, err)

	// Calculate work score as defined in TIP-45 for verification.
	expectedWorkScore := workScoreParameters.DataByte*iotago.WorkScore(tx.Size()) +
		workScoreParameters.Input*2 +
		workScoreParameters.ContextInput*3 +
		// Accounts for one Signature unlock.
		workScoreParameters.SignatureEd25519 +
		workScoreParameters.BlockIssuer +
		workScoreParameters.Staking +
		workScoreParameters.NativeToken*2 +
		workScoreParameters.Allotment*2

	require.Equal(t, expectedWorkScore, workScore, "work score expected: %d, actual: %d", expectedWorkScore, workScore)
}
