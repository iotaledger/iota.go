package bct

import (
	"math/bits"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

// MaxBatchSize is the maximum number of Curl-P hashes that can be computed in one batch.
const MaxBatchSize = bits.UintSize

// ErrInvalidBatchSize is returned when the batch size is invalid.
var ErrInvalidBatchSize = errors.New("invalid batch size")

type state struct {
	l, h      [curl.StateSize]uint
	rounds    curl.CurlRounds
	direction curl.SpongeDirection
}

func New81() *state {
	c := &state{
		rounds:    curl.CurlP81,
		direction: curl.SpongeAbsorbing,
	}
	c.Reset()
	return c
}

// Reset the internal state of the Curl sponge.
func (c *state) Reset() {
	for i := 0; i < curl.StateSize; i++ {
		c.l[i] = ^uint(0)
		c.h[i] = ^uint(0)
	}
	c.direction = curl.SpongeAbsorbing
}

// Clone returns a deep copy of the current Curl instance.
func (c *state) Clone() *state {
	return &state{
		l:         c.l,
		h:         c.h,
		rounds:    c.rounds,
		direction: c.direction,
	}
}

// Absorb fills the states of the batched sponge with src; each element of src must have the length tritsCount.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *state) Absorb(src []trinary.Trits, tritsCount int) error {
	if len(src) < 1 || len(src) > MaxBatchSize {
		return ErrInvalidBatchSize
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

// Squeeze squeezes out trits of the length tritsCount.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *state) Squeeze(dst []trinary.Trits, tritsCount int) error {
	if len(dst) < 1 || len(dst) > MaxBatchSize {
		return ErrInvalidBatchSize
	}
	if tritsCount%consts.HashTrinarySize != 0 {
		return consts.ErrInvalidSqueezeLength
	}

	for i := 0; i < tritsCount/consts.HashTrinarySize; i++ {
		// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
		if c.direction == curl.SpongeSqueezing {
			c.transform()
		}
		c.direction = curl.SpongeSqueezing
		for j := range dst {
			dst[j] = make(trinary.Trits, tritsCount)
			c.out(dst[j][i*consts.HashTrinarySize:], uint(j))
		}
	}
	return nil
}

// in sets the idx-th entry of the internal state to src.
func (c *state) in(src trinary.Trits, idx uint) {
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
func (c *state) out(dst trinary.Trits, idx uint) {
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
func (c *state) transform() {
	var ltmp, htmp [curl.StateSize]uint
	transform(&ltmp, &htmp, &c.l, &c.h, int(c.rounds))
	// for odd number of rounds we need to copy the buffer into the state
	if c.rounds%2 != 0 {
		copy(c.l[:], ltmp[:])
		copy(c.h[:], htmp[:])
	}
}

func transform(lto, hto, lfrom, hfrom *[curl.StateSize]uint, rounds int) {
	for i := 0; i < rounds; i++ {
		// three Curl-P rounds unrolled
		for j := 0; j < curl.StateSize-2; j += 3 {
			t0 := curl.Indices[j+0]
			t1 := curl.Indices[j+1]
			t2 := curl.Indices[j+2]
			t3 := curl.Indices[j+3]

			l0 := lfrom[t0]
			l1 := lfrom[t1]
			l2 := lfrom[t2]
			l3 := lfrom[t3]
			h0 := hfrom[t0]
			h1 := hfrom[t1]
			h2 := hfrom[t2]
			h3 := hfrom[t3]

			v0 := l0 & (l1 ^ h0)
			v1 := l1 & (l2 ^ h1)
			v2 := l2 & (l3 ^ h2)

			lto[j+0] = ^v0
			lto[j+1] = ^v1
			lto[j+2] = ^v2
			hto[j+0] = (l0 ^ h1) | v0
			hto[j+1] = (l1 ^ h2) | v1
			hto[j+2] = (l2 ^ h3) | v2
		}
		// swap buffers
		lfrom, lto = lto, lfrom
		hfrom, hto = hto, hfrom
	}
}
