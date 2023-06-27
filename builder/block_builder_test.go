package builder_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestBlockBuilder(t *testing.T) {
	parents := tpkg.SortedRandBlockIDs(4)

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte("hello world"),
		Data: []byte{1, 2, 3, 4},
	}
	block, err := builder.NewBlockBuilder(iotago.BlockTypeBasic).
		Payload(taggedDataPayload).
		StrongParents(parents).
		BurnedMana(100).
		Build()
	require.NoError(t, err)

	require.Equal(t, iotago.BlockTypeBasic, block.Block.Type())

	basicBlock := block.Block.(*iotago.BasicBlock)
	require.EqualValues(t, 100, basicBlock.BurnedMana)
}

// TODO: add test for validator block and panic if wrong combination is used
