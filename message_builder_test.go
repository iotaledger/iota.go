package iota_test

import (
	"context"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/require"
)

func TestMessageBuilder(t *testing.T) {
	const targetPoWScore float64 = 4000

	parent1 := rand32ByteHash()
	parent2 := rand32ByteHash()
	parent3 := rand32ByteHash()
	parent4 := rand32ByteHash()

	msg, err := iota.NewMessageBuilder().
		Payload(&iota.Indexation{Index: "hello world", Data: []byte{1, 2, 3, 4}}).
		Parents([][]byte{parent1[:], parent2[:], parent3[:], parent4[:]}).
		ProofOfWork(context.Background(), targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, err := msg.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)
}
