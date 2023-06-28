package iotago_test

import (
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestBlock_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - no payload",
			source: tpkg.RandBlock(1337),
			target: &iotago.Block{},
		},
		{
			name:   "ok - transaction",
			source: tpkg.RandBlock(iotago.PayloadTransaction),
			target: &iotago.Block{},
		},
		{
			name:   "ok - milestone",
			source: tpkg.RandBlock(iotago.PayloadMilestone),
			target: &iotago.Block{},
		},
		{
			name:   "ok - tagged data",
			source: tpkg.RandBlock(iotago.PayloadTaggedData),
			target: &iotago.Block{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestBlock_MinSize(t *testing.T) {

	block := &iotago.Block{
		ProtocolVersion: tpkg.TestAPI.ProtocolParameters().Version(),
		StrongParents:   tpkg.SortedRandBlockIDs(1),
		SlotCommitment:  iotago.NewEmptyCommitment(),
		Signature:       tpkg.RandEd25519Signature(),
		Payload:         nil,
	}

	msgBytes, err := tpkg.TestAPI.Encode(block)
	require.NoError(t, err)

	block2 := &iotago.Block{}
	_, err = tpkg.TestAPI.Decode(msgBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, block, block2)
}

func TestBlock_ProtocolVersionSyntactical(t *testing.T) {

	block := &iotago.Block{
		ProtocolVersion: tpkg.TestAPI.ProtocolParameters().Version() + 1,
		StrongParents:   tpkg.SortedRandBlockIDs(1),
		Payload:         nil,
	}

	_, err := tpkg.TestAPI.Encode(block, serix.WithValidation())
	require.Error(t, err)
}

func TestBlock_DeserializationNotEnoughData(t *testing.T) {

	blockBytes := []byte{tpkg.TestAPI.ProtocolParameters().Version(), 1}

	block := &iotago.Block{}
	_, err := tpkg.TestAPI.Decode(blockBytes, block)
	require.ErrorIs(t, err, serializer.ErrDeserializationNotEnoughData)
}
