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
	p         [3]uint256      // positive part of the state in 3 chunks of 243 trits
	n         [3]uint256      // negative part of the state in 3 chunks of 243 trits
	direction SpongeDirection // whether the sponge is absorbing or squeezing
}

// NewCurlP81 returns a new Curl-P-81.
func NewCurlP81() SpongeFunction {
	return &Curl{
		direction: SpongeAbsorbing,
	}
}

// CopyState copy the content of the Curl state buffer into s.
func (c *Curl) CopyState(s Trits) {
	for i := 0; i < 3; i++ {
		_ = s[HashTrinarySize-1]
		for j := uint(0); j <= HashTrinarySize-1; j++ {
			if c.p[i].bit(j) != 0 {
				s[j] = 1
			} else if c.n[i].bit(j) != 0 {
				s[j] = -1
			} else {
				s[i] = 0
			}
		}
		s = s[HashTrinarySize:]
	}
}

// Squeeze squeezes out trits of the given length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) Squeeze(tritsCount int) (Trits, error) {
	if tritsCount%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	out := make(Trits, tritsCount)
	for b := out; len(b) >= HashTrinarySize; b = b[HashTrinarySize:] {
		c.squeeze(b)
	}
	return out, nil
}

func (c *Curl) squeeze(hash Trits) {
	// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
	if c.direction == SpongeSqueezing {
		c.transform()
	}
	c.direction = SpongeSqueezing

	_ = hash[HashTrinarySize-1]
	for i := uint(0); i <= HashTrinarySize-1; i++ {
		if c.p[0].bit(i) != 0 {
			hash[i] = 1
		} else if c.n[0].bit(i) != 0 {
			hash[i] = -1
		}
	}
}

// transform transforms the sponge.
func (c *Curl) transform() {
	transform(&c.p, &c.n)
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
		var p, n uint256
		for i := uint(0); i < HashTrinarySize; i++ {
			switch in[i] {
			case 1:
				p.setBit(i)
			case -1:
				n.setBit(i)
			}
		}
		// the input only replaces the first 243 trits of the state
		c.p[0], c.n[0] = p, n
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

// Reset the internal state of the Curl sponge.
func (c *Curl) Reset() {
	*c = Curl{
		direction: SpongeAbsorbing,
	}
}

// Clone returns a deep copy of the current Curl.
func (c *Curl) Clone() SpongeFunction {
	return &Curl{
		p:         c.p,
		n:         c.n,
		direction: c.direction,
	}
}

// HashTrits returns the hash of the given trits.
func HashTrits(trits Trits) (Trits, error) {
	c := NewCurlP81()
	if err := c.Absorb(trits); err != nil {
		return nil, err
	}
	return c.Squeeze(HashTrinarySize)
}

// HashTrytes returns the hash of the given trytes.
func HashTrytes(t Trytes) (Trytes, error) {
	c := NewCurlP81()
	if err := c.AbsorbTrytes(t); err != nil {
		return "", err
	}
	return c.SqueezeTrytes(HashTrinarySize)
}

// MustHashTrytes returns the hash of the given trytes.
// It panics if the given trytes are not valid.
func MustHashTrytes(t Trytes) Trytes {
	trytes, err := HashTrytes(t)
	if err != nil {
		panic(err)
	}
	return trytes
}
