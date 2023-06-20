package builder_test

import (
	"context"
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
	const targetPoWScore float64 = 500

	parents := tpkg.SortedRandBlockIDs(4)

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte("hello world"),
		Data: []byte{1, 2, 3, 4},
	}
	block, err := builder.NewBlockBuilder().
		Payload(taggedDataPayload).
		StrongParents(parents).
		BurnedMana(100).
		ProofOfWork(context.Background(), targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, _, err := block.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)

	require.EqualValues(t, 100, block.BurnedMana)
}
