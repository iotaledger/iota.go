package iotago_test

import (
	"context"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/require"
)

func TestMessageBuilder(t *testing.T) {
	const targetPoWScore float64 = 500

	parents := tpkg.SortedRand32BytArray(4)

	indexationPayload := &iotago.Indexation{
		Index: []byte("hello world"),
		Data:  []byte{1, 2, 3, 4},
	}
	msg, err := iotago.NewMessageBuilder().
		Payload(indexationPayload).
		ParentsMessageIDs(parents).
		ProofOfWork(context.Background(), targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, err := msg.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)
}
