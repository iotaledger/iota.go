package iota_test

import (
	"context"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/require"
)

func TestMessageBuilder(t *testing.T) {
	const targetPoWScore float64 = 500

	parents := sortedRand32ByteHashes(4)

	msg, err := iota.NewMessageBuilder().
		Payload(&iota.Indexation{Index: "hello world", Data: []byte{1, 2, 3, 4}}).
		Parents([][]byte{parents[0][:], parents[1][:], parents[2][:], parents[3][:]}).
		ProofOfWork(context.Background(), targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, err := msg.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)
}
