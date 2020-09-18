// Package bct implements the BCT Curl hashing function computing multiple Curl hashes in parallel.
package bct

import (
	"math/bits"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/curl"
	"github.com/iotaledger/iota.go/legacy/trinary"
)

// MaxBatchSize is the maximum number of Curl hashes that can be computed in one batch.
const MaxBatchSize = bits.UintSize

type state struct {
	l, h      [curl.StateSize]uint // main batched state of the hash
	rounds    curl.CurlRounds      // number of rounds used
	direction curl.SpongeDirection // whether the sponge is absorbing or squeezing
}

// NewCurl initializes a new BCT Curl instance.
func NewCurl(rounds ...curl.CurlRounds) *state {
	r := curl.NumberOfRounds
	if len(rounds) > 0 {
		r = rounds[0]
	}
	c := &state{
		rounds:    r,
		direction: curl.SpongeAbsorbing,
	}
	c.Reset()
	return c
}

// NewCurlP27 returns a new BCT Curl-P-27.
func NewCurlP27() *state {
	return NewCurl(curl.CurlP27)
}

// NewCurlP81 returns a new BCT Curl-P-81.
func NewCurlP81() *state {
	return NewCurl(curl.CurlP81)
}

// NumRounds returns the number of rounds for the BCT Curl instance.
func (c *state) NumRounds() int {
	return int(c.rounds)
}

// Reset the internal state of the BCT Curl instance.
func (c *state) Reset() {
	for i := 0; i < curl.StateSize; i++ {
		c.l[i] = ^uint(0)
		c.h[i] = ^uint(0)
	}
	c.direction = curl.SpongeAbsorbing
}

// Clone returns a deep copy of the current BCT Curl instance.
func (c *state) Clone() *state {
	return &state{
		l:         c.l,
		h:         c.h,
		rounds:    c.rounds,
		direction: c.direction,
	}
}

// Absorb fills the states of the sponge with src; each element of src must have the length tritsCount.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *state) Absorb(src []trinary.Trits, tritsCount int) error {
	if len(src) < 1 || len(src) > MaxBatchSize {
		return legacy.ErrInvalidBatchSize
	}
	if tritsCount%legacy.HashTrinarySize != 0 {
		return legacy.ErrInvalidTritsLength
	}

	if c.direction != curl.SpongeAbsorbing {
		panic("absorb after squeeze")
	}
	for i := 0; i < tritsCount/legacy.HashTrinarySize; i++ {
		for j := range src {
			c.in(src[j][i*legacy.HashTrinarySize:], uint(j))
		}
		c.transform()
	}
	return nil
}

// Squeeze squeezes out trits of the given length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *state) Squeeze(dst []trinary.Trits, tritsCount int) error {
	if len(dst) < 1 || len(dst) > MaxBatchSize {
		return legacy.ErrInvalidBatchSize
	}
	if tritsCount%legacy.HashTrinarySize != 0 {
		return legacy.ErrInvalidSqueezeLength
	}

	for j := range dst {
		dst[j] = make(trinary.Trits, tritsCount)
	}
	for i := 0; i < tritsCount/legacy.HashTrinarySize; i++ {
		// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
		if c.direction == curl.SpongeSqueezing {
			c.transform()
		}
		c.direction = curl.SpongeSqueezing
		for j := range dst {
			c.out(dst[j][i*legacy.HashTrinarySize:], uint(j))
		}
	}
	return nil
}

// in sets the idx-th entry of the internal state to src.
func (c *state) in(src trinary.Trits, idx uint) {
	// bounds check hint to compiler
	if len(src) < legacy.HashTrinarySize {
		panic(legacy.ErrInvalidTritsLength)
	}

	s := uint(1) << idx
	u := ^s
	for i := 0; i < legacy.HashTrinarySize; i++ {
		switch src[i] {
		case 1:
			c.l[i] &= u
			c.h[i] |= s
		case -1:
			c.l[i] |= s
			c.h[i] &= u
		default:
			c.l[i] |= s
			c.h[i] |= s
		}
	}
}

// out extracts the idx-th entry of the internal state to dst.
func (c *state) out(dst trinary.Trits, idx uint) {
	// bounds check hint to compiler
	if len(dst) < legacy.HashTrinarySize {
		panic(legacy.ErrInvalidTritsLength)
	}

	for i := 0; i < legacy.HashTrinarySize; i++ {
		l := (c.l[i] >> idx) & 1
		h := (c.h[i] >> idx) & 1

		switch {
		case l == 0 && h == 1:
			dst[i] = 1
		case l == 1 && h == 0:
			dst[i] = -1
		default:
			dst[i] = 0
		}
	}
}

// transform transforms the sponge.
func (c *state) transform() {
	var ltmp, htmp [curl.StateSize]uint
	transform(&ltmp, &htmp, &c.l, &c.h, uint(c.rounds))
	// for odd number of rounds we need to copy the buffer into the state
	if c.rounds%2 != 0 {
		copy(c.l[:], ltmp[:])
		copy(c.h[:], htmp[:])
	}
}
