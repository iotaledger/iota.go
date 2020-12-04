package bct_test

import (
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/curl"
	"github.com/iotaledger/iota.go/legacy/curl/bct"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/stretchr/testify/assert"
)

func TestBCTCurl(t *testing.T) {
	type test struct {
		name    string
		src     []trinary.Trits
		hashLen int
	}

	tests := []test{
		{
			name:    "Curl-P-81: trits and hash",
			src:     Trits(bct.MaxBatchSize, legacy.HashTrinarySize),
			hashLen: legacy.HashTrinarySize,
		},
		{
			name:    "Curl-P-81: multi trits and hash",
			src:     Trits(bct.MaxBatchSize, legacy.TransactionTrinarySize),
			hashLen: legacy.HashTrinarySize,
		},
		{
			name:    "Curl-P-81: trits and multi squeeze",
			src:     Trits(bct.MaxBatchSize, legacy.HashTrinarySize),
			hashLen: 3 * legacy.HashTrinarySize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := bct.NewCurlP81()
			err := c.Absorb(tt.src, len(tt.src[0]))
			assert.NoError(t, err)

			dst := make([]trinary.Trits, len(tt.src))
			err = c.Squeeze(dst, tt.hashLen)
			assert.NoError(t, err)

			for i := range dst {
				// compare against the non batched Curl
				assert.Equal(t, dst[i], CurlSum(tt.src[i], tt.hashLen))
			}
		})
	}
}

func TestBCTCurl_Reset(t *testing.T) {
	a := []trinary.Trits{trinary.MustTrytesToTrits(strings.Repeat("A", legacy.HashTrytesSize))}
	b := []trinary.Trits{trinary.MustTrytesToTrits(strings.Repeat("B", legacy.HashTrytesSize))}

	c1 := bct.NewCurlP81()
	assert.NoError(t, c1.Absorb(a, len(a[0])))
	assert.NoError(t, c1.Squeeze(make([]trinary.Trits, 1), legacy.HashTrinarySize))

	c1.Reset()
	c2 := bct.NewCurlP81()

	assert.NoError(t, c1.Absorb(b, len(b[0])))
	assert.NoError(t, c2.Absorb(b, len(b[0])))

	hash1 := make([]trinary.Trits, 1)
	assert.NoError(t, c1.Squeeze(hash1, legacy.HashTrinarySize))
	hash2 := make([]trinary.Trits, 1)
	assert.NoError(t, c2.Squeeze(hash2, legacy.HashTrinarySize))

	assert.EqualValues(t, hash2[0], hash1[0])
}

func TestBCTCurlAbsorbAfterSqueeze(t *testing.T) {
	a := []trinary.Trits{trinary.MustTrytesToTrits(strings.Repeat("A", legacy.HashTrytesSize))}

	c := bct.NewCurlP81()
	assert.NoError(t, c.Absorb(a, len(a[0])))
	assert.NoError(t, c.Squeeze(make([]trinary.Trits, 1), legacy.HashTrinarySize))
	assert.Panics(t, func() {
		_ = c.Absorb(a, len(a[0]))
	})
}

func TestBCTCurlCLone(t *testing.T) {
	a := []trinary.Trits{trinary.MustTrytesToTrits(strings.Repeat("A", legacy.HashTrytesSize))}
	b := []trinary.Trits{trinary.MustTrytesToTrits(strings.Repeat("B", legacy.HashTrytesSize))}

	c1 := bct.NewCurlP81()
	err := c1.Absorb(a, len(a[0]))
	assert.NoError(t, err)

	c2 := c1.Clone()
	err = c1.Absorb(b, len(b[0]))
	assert.NoError(t, err)
	err = c2.Absorb(b, len(b[0]))
	assert.NoError(t, err)

	hash1 := make([]trinary.Trits, 1)
	err = c1.Squeeze(hash1, legacy.HashTrinarySize)
	assert.NoError(t, err)
	hash2 := make([]trinary.Trits, 1)
	err = c2.Squeeze(hash2, legacy.HashTrinarySize)
	assert.NoError(t, err)

	assert.Equal(t, hash2[0], hash1[0])
}

func Trits(size int, tritsCount int) []trinary.Trits {
	trytesCount := tritsCount / legacy.TritsPerTryte
	src := make([]trinary.Trits, size)
	for i := range src {
		trytes := strings.Repeat("ABC", trytesCount/3+1)[:trytesCount-2] + trinary.IntToTrytes(int64(i), 2)
		src[i] = trinary.MustTrytesToTrits(trytes)
	}
	return src
}

func CurlSum(data trinary.Trits, tritsCount int) trinary.Trits {
	c := curl.NewCurlP81()
	if err := c.Absorb(data); err != nil {
		panic(err)
	}
	out, err := c.Squeeze(tritsCount)
	if err != nil {
		panic(err)
	}
	return out
}
