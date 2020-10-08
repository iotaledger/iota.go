// Package curl implements the Curl hashing function.
package curl

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

// Curl is a sponge function with an internal state of size StateSize.
// b = r + c, b = StateSize, r = HashSize, c = StateSize - HashSize
type Curl struct {
	state     [StateSize]int8 // main state of the hash
	rounds    CurlRounds      // number of rounds used
	direction SpongeDirection // whether the sponge is absorbing or squeezing
}

// NewCurl initializes a new Curl instance.
func NewCurl(rounds ...CurlRounds) SpongeFunction {
	r := NumberOfRounds
	if len(rounds) > 0 {
		r = rounds[0]
	}
	return &Curl{
		rounds:    r,
		direction: SpongeAbsorbing,
	}
}

// NewCurlP27 returns a new Curl-P-27.
func NewCurlP27() SpongeFunction {
	return NewCurl(CurlP27)
}

// NewCurlP81 returns a new Curl-P-81.
func NewCurlP81() SpongeFunction {
	return NewCurl(CurlP81)
}

// NumRounds returns the number of rounds for this Curl instance.
func (c *Curl) NumRounds() int {
	return int(c.rounds)
}

// CopyState copy the content of the Curl state buffer into s.
func (c *Curl) CopyState(s Trits) {
	copy(s, c.state[:])
}

// Squeeze squeezes out trits of the given length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) Squeeze(tritsCount int) (Trits, error) {
	if tritsCount%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	out := make(Trits, tritsCount)
	for p := out; len(p) >= HashTrinarySize; p = p[HashTrinarySize:] {
		c.squeeze(p)
	}
	return out, nil
}

func (c *Curl) squeeze(hash Trits) {
	// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
	if c.direction == SpongeSqueezing {
		c.transform()
	}
	c.direction = SpongeSqueezing
	copy(hash, c.state[:HashTrinarySize])
}

// MustSqueeze squeezes out trits of the given trit length.
// The value tritsCount has to be a multiple of HashTrinarySize. It panics if the length is not valid.
func (c *Curl) MustSqueeze(tritsCount int) Trits {
	out, err := c.Squeeze(tritsCount)
	if err != nil {
		panic(err)
	}
	return out
}

// SqueezeTrytes squeezes out trytes of the given trit length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) SqueezeTrytes(tritsCount int) (Trytes, error) {
	trits, err := c.Squeeze(tritsCount)
	if err != nil {
		return "", err
	}
	return MustTritsToTrytes(trits), nil
}

// MustSqueezeTrytes squeezes out trytes of the given trit length.
// The value tritsCount has to be a multiple of HashTrinarySize. It panics if the length is not valid.
func (c *Curl) MustSqueezeTrytes(tritsCount int) Trytes {
	return MustTritsToTrytes(c.MustSqueeze(tritsCount))
}

// Absorb fills the internal state of the sponge with the given trits.
func (c *Curl) Absorb(in Trits) error {
	if len(in) == 0 || len(in)%HashTrinarySize != 0 {
		return errors.Wrap(ErrInvalidTritsLength, "trits slice length must be a multiple of 243")
	}

	if c.direction != SpongeAbsorbing {
		panic("absorb after squeeze")
	}
	for len(in) >= HashTrinarySize {
		copy(c.state[:HashTrinarySize], in)
		in = in[HashTrinarySize:]

		c.transform()
	}
	return nil
}

// AbsorbTrytes fills the internal state of the sponge with the given trytes.
func (c *Curl) AbsorbTrytes(in Trytes) error {
	if len(in) == 0 || len(in)%HashTrytesSize != 0 {
		return errors.Wrap(ErrInvalidTrytesLength, "trytes length must be a multiple of 81")
	}

	trits, err := TrytesToTrits(in)
	if err != nil {
		return err
	}
	return c.Absorb(trits)
}

// AbsorbTrytes fills the internal state of the sponge with the given trytes.
// It panics if the given trytes are not valid.
func (c *Curl) MustAbsorbTrytes(in Trytes) {
	err := c.Absorb(MustTrytesToTrits(in))
	if err != nil {
		panic(err)
	}
}

// transform transforms the sponge.
func (c *Curl) transform() {
	var tmp [StateSize]int8
	transform(&tmp, &c.state, uint(c.rounds))
	// for odd number of rounds we need to copy the buffer into the state
	if c.rounds%2 != 0 {
		copy(c.state[:], tmp[:])
	}
}

// Reset the internal state of the Curl sponge.
func (c *Curl) Reset() {
	for i := range c.state {
		c.state[i] = 0
	}
	c.direction = SpongeAbsorbing
}

// HashTrits returns the hash of the given trits.
func HashTrits(trits Trits, rounds ...CurlRounds) (Trits, error) {
	c := NewCurl(rounds...)
	if err := c.Absorb(trits); err != nil {
		return nil, err
	}
	return c.Squeeze(HashTrinarySize)
}

// HashTrytes returns the hash of the given trytes.
func HashTrytes(t Trytes, rounds ...CurlRounds) (Trytes, error) {
	c := NewCurl(rounds...)
	if err := c.AbsorbTrytes(t); err != nil {
		return "", err
	}
	return c.SqueezeTrytes(HashTrinarySize)
}

// MustHashTrytes returns the hash of the given trytes.
// It panics if the given trytes are not valid.
func MustHashTrytes(t Trytes, rounds ...CurlRounds) Trytes {
	trytes, err := HashTrytes(t, rounds...)
	if err != nil {
		panic(err)
	}
	return trytes
}

// Clone returns a deep copy of the current Curl
func (c *Curl) Clone() SpongeFunction {
	return &Curl{
		state:     c.state,
		rounds:    c.rounds,
		direction: c.direction,
	}
}
