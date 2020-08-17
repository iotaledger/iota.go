// Package curl implements the Curl hashing function.
package curl

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

// CurlRounds is the default number of rounds used in transform.
type CurlRounds int

const (
	// StateSize is the size of the Curl hash function.
	StateSize = HashTrinarySize * 3

	// CurlP27 is used for hashing with 27 rounds
	CurlP27 CurlRounds = 27

	// CurlP81 is used for hashing with 81 rounds
	CurlP81 CurlRounds = 81

	// NumberOfRounds is the default number of rounds in transform.
	NumberOfRounds = CurlP81
)

// spongeDirection indicates the direction trits are flowing through the sponge.
type spongeDirection int

const (
	// spongeAbsorbing indicates that the sponge is absorbing input.
	spongeAbsorbing spongeDirection = iota
	// spongeSqueezing indicates that the sponge is being squeezed.
	spongeSqueezing
)

var (
	// TruthTable of the Curl hash function.
	TruthTable = [11]int8{1, 0, -1, 2, 1, -1, 0, 2, -1, 1, 0}
	// Indices of the Curl hash function.
	Indices [StateSize + 1]int
)

func init() {
	for i := 0; i < StateSize; i++ {
		p := -365

		if Indices[i] < 365 {
			p = 364
		}

		Indices[i+1] = Indices[i] + p
	}
}

// Curl is a sponge function with an internal state of size StateSize.
// b = r + c, b = StateSize, r = HashSize, c = StateSize - HashSize
type Curl struct {
	state  [StateSize]int8
	rounds CurlRounds
	mode   spongeDirection
}

// NewCurl initializes a new instance with an empty state.
func NewCurl(rounds ...CurlRounds) SpongeFunction {
	curlRounds := NumberOfRounds
	if len(rounds) > 0 {
		curlRounds = rounds[0]
	}
	return &Curl{
		rounds: curlRounds,
		mode:   spongeAbsorbing,
	}
}

// NewCurlP27 returns a new CurlP27.
func NewCurlP27() SpongeFunction {
	return NewCurl(CurlP27)
}

// NewCurlP81 returns a new CurlP81.
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

// Squeeze squeezes out trits of the given length. Length has to be a multiple of HashTrinarySize.
func (c *Curl) Squeeze(length int) (Trits, error) {
	if length%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	out := make(Trits, length)
	for p := out; len(p) >= HashTrinarySize; p = p[HashTrinarySize:] {
		c.squeeze(p)
	}
	return out, nil
}

func (c *Curl) squeeze(hash Trits) {
	// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
	if c.mode == spongeSqueezing {
		c.transform()
	}
	copy(hash, c.state[:HashTrinarySize])
	c.mode = spongeSqueezing
}

// MustSqueeze squeezes out trits of the given length. Length has to be a multiple of HashTrinarySize.
// It panics if the length is not valid.
func (c *Curl) MustSqueeze(length int) Trits {
	out, err := c.Squeeze(length)
	if err != nil {
		panic(err)
	}
	return out
}

// SqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
func (c *Curl) SqueezeTrytes(length int) (Trytes, error) {
	trits, err := c.Squeeze(length)
	if err != nil {
		return "", err
	}
	return MustTritsToTrytes(trits), nil
}

// MustSqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
// It panics if the trytes or the length are not valid.
func (c *Curl) MustSqueezeTrytes(length int) Trytes {
	return MustTritsToTrytes(c.MustSqueeze(length))
}

// Absorb fills the internal state of the sponge with the given trits.
func (c *Curl) Absorb(in Trits) error {
	if len(in) == 0 || len(in)%HashTrinarySize != 0 {
		return errors.Wrap(ErrInvalidTritsLength, "trits slice length must be a multiple of 243")
	}

	if c.mode != spongeAbsorbing {
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

// transform the sponge func.
func (c *Curl) transform() {
	var tmp [StateSize]int8
	transform(&tmp, &c.state, c.NumRounds())
	// for odd number of rounds we need to copy the buffer into the state
	if c.rounds%2 != 0 {
		copy(c.state[:], tmp[:])
	}
}

// Reset the internal state of the Curl sponge by filling it with all 0's.
func (c *Curl) Reset() {
	for i := range c.state {
		c.state[i] = 0
	}
	c.mode = spongeAbsorbing
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
		state:  c.state,
		rounds: c.rounds,
		mode:   c.mode,
	}
}
