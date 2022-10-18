package builder_test

import (
	"context"
	"math/rand"

	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

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
		Parents(parents).
		ProofOfWork(context.Background(), targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, _, err := block.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)
}
