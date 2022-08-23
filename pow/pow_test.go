package pow

import (
	"context"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"

	legacy "github.com/iotaledger/iota.go/consts"
)

const (
	workers     = 2
	targetScore = 4000.
)

var testWorker = New(workers)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	// call the tests
	os.Exit(m.Run())
}

func TestScore(t *testing.T) {
	tests := []*struct {
		msg    []byte
		expPoW float64
	}{
		{msg: []byte{0, 0, 0, 0, 0, 0, 0, 0}, expPoW: math.Pow(3, 1) / 8},
		{msg: []byte{203, 124, 2, 0, 0, 0, 0, 0}, expPoW: math.Pow(3, 10) / 8},
		{msg: []byte{65, 235, 119, 85, 85, 85, 85, 85}, expPoW: math.Pow(3, 14) / 8},
		{msg: make([]byte, 10000), expPoW: math.Pow(3, 0) / 10000},
	}

	for _, tt := range tests {
		pow := Score(tt.msg)
		assert.Equal(t, tt.expPoW, pow)
	}
}

func TestWorker_Mine(t *testing.T) {
	msg := append([]byte("Hello, World!"), make([]byte, nonceBytes)...)
	nonce, err := testWorker.Mine(context.Background(), msg[:len(msg)-nonceBytes], targetScore)
	require.NoError(t, err)

	// add nonce to block and check the resulting PoW score
	binary.LittleEndian.PutUint64(msg[len(msg)-nonceBytes:], nonce)
	pow := Score(msg)
	assert.GreaterOrEqual(t, pow, targetScore)
}

func TestWorker_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	go func() {
		_, err = testWorker.Mine(ctx, nil, math.MaxInt32)
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()

	assert.Eventually(t, func() bool { return errors.Is(err, ErrCancelled) }, time.Second, 10*time.Millisecond)
}

const benchBytesLen = 1600

func BenchmarkScore(b *testing.B) {
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = make([]byte, benchBytesLen)
		//nolint:gosec // we do not care about weak random numbers here
		if _, err := rand.Read(data[i]); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()

	for i := range data {
		_ = Score(data[i])
	}
}

func BenchmarkWorker(b *testing.B) {
	var (
		wg      sync.WaitGroup
		w       = New(1)
		digest  = blake2b.Sum256(nil)
		done    uint32
		counter uint64
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = w.worker(digest[:], 0, legacy.HashTrinarySize, &done, &counter)
	}()
	b.ResetTimer()
	for atomic.LoadUint64(&counter) < uint64(b.N) {
	}
	atomic.StoreUint32(&done, 1)
	wg.Wait()
}
