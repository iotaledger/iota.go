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

	msg, err := iota.NewMessageBuilder().
		Payload(&iota.Indexation{Index: "hello world", Data: []byte{1, 2, 3, 4}}).
		Parent1(parent1[:]).
		Parent2(parent2[:]).
		ProofOfWork(context.Background(), targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, err := msg.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)
}
