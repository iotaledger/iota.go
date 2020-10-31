package pow

import (
	"context"
	"crypto"
	"encoding/binary"
	"math"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	workers     = 2
	targetScore = 4000.
)

var testWorker = New(workers)

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

	// add nonce to message and check the resulting PoW score
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

	assert.Eventually(t, func() bool { return err == ErrCancelled }, time.Second, 10*time.Millisecond)
}

const benchBytesLen = 1600

func BenchmarkScore(b *testing.B) {
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = make([]byte, benchBytesLen)
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
		w       = New(1)
		digest  = make([]byte, crypto.BLAKE2b_256.Size())
		done    uint32
		counter uint64
	)
	go func() {
		_, _ = w.worker(digest[:], 0, math.MaxInt32, &done, &counter)
	}()
	b.ResetTimer()
	for atomic.LoadUint64(&counter) < uint64(b.N) {
	}
	atomic.StoreUint32(&done, 1)
}
