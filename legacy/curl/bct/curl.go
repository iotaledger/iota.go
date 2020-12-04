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

// Curl is the BCT version of the Curl hashing function.
type Curl struct {
	l, h      [curl.StateSize]uint // main batched state of the hash
	direction curl.SpongeDirection // whether the sponge is absorbing or squeezing
}

// NewCurlP81 returns a new BCT Curl-P-81.
func NewCurlP81() *Curl {
	c := &Curl{}
	c.Reset()
	return c
}

// Reset the internal state of the BCT Curl instance.
func (c *Curl) Reset() {
	for i := 0; i < curl.StateSize; i++ {
		c.l[i], c.h[i] = ^uint(0), ^uint(0)
	}
	c.direction = curl.SpongeAbsorbing
}

// Clone returns a deep copy of the current BCT Curl instance.
func (c *Curl) Clone() *Curl {
	return &Curl{
		l:         c.l,
		h:         c.h,
		direction: c.direction,
	}
}

// Absorb fills the states of the sponge with src; each element of src must have the length tritsCount.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) Absorb(src []trinary.Trits, tritsCount int) error {
	if len(src) < 1 || len(src) > MaxBatchSize {
		return legacy.ErrInvalidBatchSize
	}
	if tritsCount%legacy.HashTrinarySize != 0 {
		return legacy.ErrInvalidTritsLength
	}

	if c.direction != curl.SpongeAbsorbing {
		panic("absorb after squeeze")
	}
	for i := 0; i < tritsCount; i += legacy.HashTrinarySize {
		// reset the first 243 trits of the state, since they will be overridden by c.in
		for j := 0; j < legacy.HashTrinarySize; j++ {
			c.l[j], c.h[j] = ^uint(0), ^uint(0)
		}
		for j := range src {
			c.in(src[j][i:], uint(j))
		}
		c.transform()
	}
	return nil
}

// Squeeze squeezes out trits of the given length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) Squeeze(dst []trinary.Trits, tritsCount int) error {
	if len(dst) < 1 || len(dst) > MaxBatchSize {
		return legacy.ErrInvalidBatchSize
	}
	if tritsCount%legacy.HashTrinarySize != 0 {
		return legacy.ErrInvalidSqueezeLength
	}

	for j := range dst {
		dst[j] = make(trinary.Trits, tritsCount)
	}
	for i := 0; i < tritsCount; i += legacy.HashTrinarySize {
		// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
		if c.direction == curl.SpongeSqueezing {
			c.transform()
		}
		c.direction = curl.SpongeSqueezing
		for j := range dst {
			c.out(dst[j][i:], uint(j))
		}
	}
	return nil
}

// in sets the idx-th entry of the internal state to src.
func (c *Curl) in(src trinary.Trits, idx uint) {
	// bounds check hint to compiler
	if len(src) < legacy.HashTrinarySize {
		panic(legacy.ErrInvalidTritsLength)
	}

	idx &= bits.UintSize - 1 // hint to the compiler that shifts don't need guard code
	m := ^(uint(1) << idx)
	for i := 0; i < legacy.HashTrinarySize; i++ {
		switch src[i] {
		case 1:
			c.l[i] &= m
		case -1:
			c.h[i] &= m
		}
	}
}

// out extracts the idx-th entry of the internal state to dst.
func (c *Curl) out(dst trinary.Trits, idx uint) {
	// bounds check hint to compiler
	if len(dst) < legacy.HashTrinarySize {
		panic(legacy.ErrInvalidTritsLength)
	}

	idx &= bits.UintSize - 1 // hint to the compiler that idx is always smaller UintSize
	m := uint(1) << idx
	for i := 0; i < legacy.HashTrinarySize; i++ {
		l, h := c.l[i]&m, c.h[i]&m
		switch {
		case l == 0:
			dst[i] = 1
		case h == 0:
			dst[i] = -1
		}
	}
}

// transform transforms the sponge.
func (c *Curl) transform() {
	var ltmp, htmp [curl.StateSize]uint
	transform(&ltmp, &htmp, &c.l, &c.h, curl.NumRounds)
	c.l, c.h = ltmp, htmp
}
