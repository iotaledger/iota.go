//#nosec G404

package pow_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/iota.go/v4/blockissuer/pow"
)

const (
	workerCount         = 2
	targetTrailingZeros = 10
)

var testWorker = pow.New(workerCount)

func TestWorker_Mine(t *testing.T) {
	msg := []byte("Hello, World!")
	nonce, err := testWorker.Mine(context.Background(), msg, targetTrailingZeros)
	require.NoError(t, err)

	// check the result
	msgDigest := blake2b.Sum256(msg)
	trailingZeros := pow.TrailingZeros(msgDigest[:], nonce)
	assert.GreaterOrEqual(t, trailingZeros, targetTrailingZeros)
}

func TestWorker_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	go func() {
		_, err = testWorker.Mine(ctx, nil, 64)
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()

	assert.Eventually(t, func() bool { return ierrors.Is(err, pow.ErrCanceled) }, time.Second, 10*time.Millisecond)
}

func BenchmarkMine(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	worker := pow.New(4)

	var err error
	go func() {
		b.ResetTimer()
		_, err = worker.Mine(ctx, []byte("testdata"), 27)
		require.NoError(b, err)
		wg.Done()
	}()
	wg.Wait()
}
