package bct

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

func TestCurl(t *testing.T) {
	src := make([]trinary.Trits, MaxBatchSize)
	for i := range src {
		src[i] = trinary.MustPadTrits(trinary.IntToTrits(int64(i)), consts.HashTrinarySize)
	}

	c := New81()
	assert.NoError(t, c.Absorb(src, consts.HashTrinarySize))

	dst := make([]trinary.Trits, MaxBatchSize)
	assert.NoError(t, c.Squeeze(dst, consts.HashTrinarySize))

	for i := range dst {
		expected, _ := curl.HashTrits(src[i], curl.CurlP81)
		assert.Equal(t, expected, dst[i])
	}
}

func BenchmarkCurlTransaction(b *testing.B) {
	src := make([][]trinary.Trits, b.N)
	for i := range src {
		src[i] = make([]trinary.Trits, MaxBatchSize)
		for j := range src[i] {
			src[i][j] = randomTrits(consts.TransactionTrinarySize)
		}
	}
	dst := make([]trinary.Trits, MaxBatchSize)
	b.ResetTimer()

	for i := range src {
		c := New81()
		c.Absorb(src[i], consts.TransactionTrinarySize)
		c.Squeeze(dst, consts.HashTrinarySize)
	}
}

func BenchmarkCurlHash(b *testing.B) {
	src := make([][]trinary.Trits, b.N)
	for i := range src {
		src[i] = make([]trinary.Trits, MaxBatchSize)
		for j := range src[i] {
			src[i][j] = randomTrits(consts.HashTrinarySize)
		}
	}
	dst := make([]trinary.Trits, MaxBatchSize)
	b.ResetTimer()

	for i := range src {
		c := New81()
		c.Absorb(src[i], consts.HashTrinarySize)
		c.Squeeze(dst, consts.HashTrinarySize)
	}
}

func randomTrits(n int) trinary.Trits {
	trytes := randomTrytes((n + 2) / 3)
	return trinary.MustTrytesToTrits(trytes)[:n]
}

func randomTrytes(n int) trinary.Trytes {
	var result strings.Builder
	result.Grow(n)
	for i := 0; i < n; i++ {
		result.WriteByte(consts.TryteAlphabet[rand.Intn(len(consts.TryteAlphabet))])
	}
	return result.String()
}
