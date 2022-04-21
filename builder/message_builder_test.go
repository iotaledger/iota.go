package builder_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	// call the tests
	os.Exit(m.Run())
}

func TestMessageBuilder(t *testing.T) {
	const targetPoWScore float64 = 500

	parents := tpkg.SortedRand32BytArray(4)

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte("hello world"),
		Data: []byte{1, 2, 3, 4},
	}
	msg, err := builder.NewMessageBuilder(tpkg.TestProtoParas.Version).
		Payload(taggedDataPayload).
		ParentsMessageIDs(parents).
		ProofOfWork(context.Background(), tpkg.TestProtoParas, targetPoWScore).
		Build()
	require.NoError(t, err)

	powScore, err := msg.POW()
	require.NoError(t, err)
	require.GreaterOrEqual(t, powScore, targetPoWScore)
}
