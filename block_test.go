package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestBlock_DeSerialize(t *testing.T) {
	// TODO: what does this test actually do?
	tests := []deSerializeTest{
		{
			name:   "ok - no payload",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(1337)),
			target: &iotago.ProtocolBlock{},
		},
		{
			name:   "ok - transaction",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(iotago.PayloadTransaction)),
			target: &iotago.ProtocolBlock{},
		},
		{
			name:   "ok - milestone",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(iotago.PayloadMilestone)),
			target: &iotago.ProtocolBlock{},
		},
		{
			name:   "ok - tagged data",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(iotago.PayloadTaggedData)),
			target: &iotago.ProtocolBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestProtocolBlock_ProtocolVersionSyntactical(t *testing.T) {
	block := &iotago.ProtocolBlock{
		ProtocolVersion: tpkg.TestProtocolVersion + 1,
		SlotCommitment:  iotago.NewEmptyCommitment(),
		Signature:       tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents: tpkg.SortedRandBlockIDs(1),
			Payload:       nil,
		},
	}

	_, err := v3API.Encode(block, serix.WithValidation())
	require.ErrorContains(t, err, "mismatched protocol version")
}

func TestProtocolBlock_DeserializationNotEnoughData(t *testing.T) {
	blockBytes := []byte{tpkg.TestProtocolVersion, 1}

	block := &iotago.ProtocolBlock{}
	_, err := v3API.Decode(blockBytes, block)
	require.ErrorIs(t, err, serializer.ErrDeserializationNotEnoughData)
}

func TestBasicBlock_MinSize(t *testing.T) {
	minProtocolBlock := &iotago.ProtocolBlock{
		ProtocolVersion: tpkg.TestProtocolVersion,
		SlotCommitment:  iotago.NewEmptyCommitment(),
		Signature:       tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents: tpkg.SortedRandBlockIDs(1),
			Payload:       nil,
		},
	}

	blockBytes, err := v3API.Encode(minProtocolBlock)
	require.NoError(t, err)

	block2 := &iotago.ProtocolBlock{}
	consumedBytes, err := v3API.Decode(blockBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, minProtocolBlock, block2)
	require.Equal(t, len(blockBytes), consumedBytes)
}

func TestValidatorBlock_MinSize(t *testing.T) {
	minProtocolBlock := &iotago.ProtocolBlock{
		ProtocolVersion: tpkg.TestProtocolVersion,
		SlotCommitment:  iotago.NewEmptyCommitment(),
		Signature:       tpkg.RandEd25519Signature(),
		Block: &iotago.ValidatorBlock{
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			HighestSupportedVersion: tpkg.TestProtocolVersion,
		},
	}

	blockBytes, err := v3API.Encode(minProtocolBlock)
	require.NoError(t, err)

	block2 := &iotago.ProtocolBlock{}
	consumedBytes, err := v3API.Decode(blockBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, minProtocolBlock, block2)
	require.Equal(t, len(blockBytes), consumedBytes)
}

func TestValidatorBlock_HighestSupportedVersion(t *testing.T) {
	protocolBlock := &iotago.ProtocolBlock{
		ProtocolVersion: tpkg.TestProtocolVersion,
		SlotCommitment:  iotago.NewEmptyCommitment(),
		Signature:       tpkg.RandEd25519Signature(),
	}

	// Invalid HighestSupportedVersion.
	{
		protocolBlock.Block = &iotago.ValidatorBlock{
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			HighestSupportedVersion: tpkg.TestProtocolVersion - 1,
		}
		blockBytes, err := v3API.Encode(protocolBlock)
		require.NoError(t, err)

		block2 := &iotago.ProtocolBlock{}
		_, err = v3API.Decode(blockBytes, block2, serix.WithValidation())
		require.ErrorContains(t, err, "highest supported version")
	}

	// Valid HighestSupportedVersion.
	{
		protocolBlock.Block = &iotago.ValidatorBlock{
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			HighestSupportedVersion: tpkg.TestProtocolVersion,
		}
		blockBytes, err := v3API.Encode(protocolBlock)
		require.NoError(t, err)

		block2 := &iotago.ProtocolBlock{}
		consumedBytes, err := v3API.Decode(blockBytes, block2, serix.WithValidation())
		require.NoError(t, err)
		require.Equal(t, protocolBlock, block2)
		require.Equal(t, len(blockBytes), consumedBytes)
	}
}

// TODO: add tests
//  - max size
//  - parents parameters basic block
//  - parents parameters validator block
