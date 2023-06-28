package builder_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestMain(m *testing.M) {
	// call the tests
	os.Exit(m.Run())
}

func TestBlockBuilder(t *testing.T) {
	parents := tpkg.SortedRandBlockIDs(4)

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte("hello world"),
		Data: []byte{1, 2, 3, 4},
	}
	block, err := builder.NewBlockBuilder(tpkg.TestAPI).
		Payload(taggedDataPayload).
		StrongParents(parents).
		BurnedMana(100).
		Build()
	require.NoError(t, err)

	require.EqualValues(t, 100, block.BurnedMana)
}
