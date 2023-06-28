package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestAttestation(t *testing.T) {
	block, err := builder.NewValidatorBlockBuilder(tpkg.TestAPI).
		StrongParents(tpkg.SortedRandBlockIDs(2)).
		Sign(tpkg.RandAccountID(), tpkg.RandEd25519PrivateKey()).
		Build()

	require.NoError(t, err)
	require.Equal(t, iotago.BlockTypeValidator, block.Block.Type())

	attestation := iotago.NewAttestation(tpkg.TestAPI, block)

	// Compare fields of block and attestation.
	{
		require.Equal(t, block.ProtocolVersion, attestation.Version)
		require.Equal(t, block.IssuerID, attestation.IssuerID)
		require.Equal(t, block.IssuingTime, attestation.IssuingTime)
		require.Equal(t, block.SlotCommitment.MustID(tpkg.TestAPI), attestation.SlotCommitmentID)
		require.Equal(t, lo.PanicOnErr(block.ContentHash(tpkg.TestAPI)), attestation.BlockContentHash)
		require.Equal(t, block.Signature, attestation.Signature)
	}

	// Compare block ID and attestation block ID.
	{
		blockID, err := block.ID(tpkg.TestAPI)
		require.NoError(t, err)

		blockIDFromAttestation, err := attestation.BlockID(tpkg.TestAPI)
		require.NoError(t, err)

		require.Equal(t, blockID, blockIDFromAttestation)
	}

	// Check validity of signature.
	{
		valid, err := attestation.VerifySignature(tpkg.TestAPI)
		require.NoError(t, err)
		require.True(t, valid)
	}
}
