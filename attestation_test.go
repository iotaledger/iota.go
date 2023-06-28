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
	iotago.SwapInternalAPI(v3API)

	block, err := builder.NewValidatorBlockBuilder().
		StrongParents(tpkg.SortedRandBlockIDs(2)).
		Sign(tpkg.RandAccountID(), tpkg.RandEd25519PrivateKey()).
		Build()

	require.NoError(t, err)
	require.Equal(t, iotago.BlockTypeValidator, block.Block.Type())

	attestation := iotago.NewAttestation(block)

	// Compare fields of block and attestation.
	{
		require.Equal(t, block.ProtocolVersion, attestation.Version)
		require.Equal(t, block.IssuerID, attestation.IssuerID)
		require.Equal(t, block.IssuingTime, attestation.IssuingTime)
		require.Equal(t, block.SlotCommitment.MustID(), attestation.SlotCommitmentID)
		require.Equal(t, lo.PanicOnErr(block.ContentHash()), attestation.BlockContentHash)
		require.Equal(t, block.Signature, attestation.Signature)
	}

	// Compare block ID and attestation block ID.
	{
		blockID, err := block.ID(v3API.TimeProvider())
		require.NoError(t, err)

		blockIDFromAttestation, err := attestation.BlockID(v3API.TimeProvider())
		require.NoError(t, err)

		require.Equal(t, blockID, blockIDFromAttestation)
	}

	// Check validity of signature.
	{
		valid, err := attestation.VerifySignature()
		require.NoError(t, err)
		require.True(t, valid)
	}
}
