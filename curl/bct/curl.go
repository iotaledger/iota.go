// Package bct implements the BCT Curl hashing function computing multiple Curl hashes in parallel.
package bct

import (
	"math/bits"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
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
	c := &Curl{
		direction: curl.SpongeAbsorbing,
	}
	c.Reset()
	return c
}

// Reset the internal state of the BCT Curl instance.
func (c *Curl) Reset() {
	for i := 0; i < curl.StateSize; i++ {
		c.l[i] = ^uint(0)
		c.h[i] = ^uint(0)
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
		return consts.ErrInvalidBatchSize
	}
	if tritsCount%consts.HashTrinarySize != 0 {
		return consts.ErrInvalidTritsLength
	}

	if c.direction != curl.SpongeAbsorbing {
		panic("absorb after squeeze")
	}
	for i := 0; i < tritsCount/consts.HashTrinarySize; i++ {
		for j := range src {
			c.in(src[j][i*consts.HashTrinarySize:], uint(j))
		}
		c.transform()
	}
	return nil
}

// Squeeze squeezes out trits of the given length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) Squeeze(dst []trinary.Trits, tritsCount int) error {
	if len(dst) < 1 || len(dst) > MaxBatchSize {
		return consts.ErrInvalidBatchSize
	}
	if tritsCount%consts.HashTrinarySize != 0 {
		return consts.ErrInvalidSqueezeLength
	}

	for j := range dst {
		dst[j] = make(trinary.Trits, tritsCount)
	}
	for i := 0; i < tritsCount/consts.HashTrinarySize; i++ {
		// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
		if c.direction == curl.SpongeSqueezing {
			c.transform()
		}
		c.direction = curl.SpongeSqueezing
		for j := range dst {
			c.out(dst[j][i*consts.HashTrinarySize:], uint(j))
		}
	}
	return nil
}

// in sets the idx-th entry of the internal state to src.
func (c *Curl) in(src trinary.Trits, idx uint) {
	// bounds check hint to compiler
	if len(src) < consts.HashTrinarySize {
		panic(consts.ErrInvalidTritsLength)
	}

	s := uint(1) << idx
	u := ^s
	for i := 0; i < consts.HashTrinarySize; i++ {
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
func (c *Curl) out(dst trinary.Trits, idx uint) {
	// bounds check hint to compiler
	if len(dst) < consts.HashTrinarySize {
		panic(consts.ErrInvalidTritsLength)
	}

	for i := 0; i < consts.HashTrinarySize; i++ {
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
func (c *Curl) transform() {
	var ltmp, htmp [curl.StateSize]uint
	transform(&ltmp, &htmp, &c.l, &c.h, curl.NumRounds)
	copy(c.l[:], ltmp[:])
	copy(c.h[:], htmp[:])
}
