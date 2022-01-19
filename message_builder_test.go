package iotago_test

import (
	"context"
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestMessageBuilder(t *testing.T) {
	const targetPoWScore float64 = 500

	parents := tpkg.SortedRand32BytArray(4)

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte("hello world"),
		Data: []byte{1, 2, 3, 4},
	}
	msg, err := iotago.NewMessageBuilder().
		Payload(taggedDataPayload).
		ParentsMessageIDs(parents).
		ProofOfWork(context.Background(), DefZeroRentParas, targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, err := msg.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)
}
