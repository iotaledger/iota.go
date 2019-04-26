// Package curl implements the Curl hashing function.
package curl

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
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

var (
	// optional transform function in C.
	transformC func(Trits, int)
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

// Curl is a sponge function with an internal State of size StateSize.
// b = r + c, b = StateSize, r = HashSize, c = StateSize - HashSize
type Curl struct {
	State  Trits
	Rounds CurlRounds
}

// NewCurl initializes a new instance with an empty State.
func NewCurl(rounds ...CurlRounds) SpongeFunction {
	curlRounds := NumberOfRounds

	if len(rounds) > 0 {
		curlRounds = rounds[0]
	}

	c := &Curl{
		State:  make(Trits, StateSize),
		Rounds: curlRounds,
	}
	return c
}

// NewCurlP27 returns a new CurlP27.
func NewCurlP27() SpongeFunction {
	return NewCurl(CurlP27)
}

// NewCurlP81 returns a new CurlP81.
func NewCurlP81() SpongeFunction {
	return NewCurl(CurlP81)
}

// Squeeze squeezes out trits of the given length. Length has to be a multiple of HashTrinarySize.
func (c *Curl) Squeeze(length int) (Trits, error) {
	if length%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	out := make(Trits, length)
	for i := 1; i <= length/HashTrinarySize; i++ {
		copy(out[HashTrinarySize*(i-1):HashTrinarySize*i], c.State[:HashTrinarySize])
		c.Transform()
	}

	return out, nil
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
	return TritsToTrytes(trits)
}

// MustSqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
// It panics if the trytes or the length are not valid.
func (c *Curl) MustSqueezeTrytes(length int) Trytes {
	return MustTritsToTrytes(c.MustSqueeze(length))
}

// Absorb fills the internal State of the sponge with the given trits.
func (c *Curl) Absorb(in Trits) error {
	var lenn int
	for i := 0; i < len(in); i += lenn {
		lenn = HashTrinarySize

		if len(in)-i < HashTrinarySize {
			lenn = len(in) - i
		}

		copy(c.State, in[i:i+lenn])
		c.Transform()
	}
	return nil
}

// AbsorbTrytes fills the internal State of the sponge with the given trytes.
func (c *Curl) AbsorbTrytes(inn Trytes) error {
	var in Trits
	var err error

	if len(inn) == 0 {
		in = Trits{0}
	} else {
		in, err = TrytesToTrits(inn)
		if err != nil {
			return err
		}
	}
	return c.Absorb(in)
}

// AbsorbTrytes fills the internal State of the sponge with the given trytes.
// It panics if the given trytes are not valid.
func (c *Curl) MustAbsorbTrytes(inn Trytes) {
	err := c.AbsorbTrytes(inn)
	if err != nil {
		panic(err)
	}
}

// Transform does Transform in sponge func.
func (c *Curl) Transform() {
	if transformC != nil {
		transformC(c.State, int(c.Rounds))
		return
	}

	var cpy [StateSize]int8

	for r := c.Rounds; r > 0; r-- {
		copy(cpy[:], c.State)
		for i := 0; i < StateSize; i++ {
			t1 := Indices[i]
			t2 := Indices[i+1]
			c.State[i] = TruthTable[cpy[t1]+(cpy[t2]<<2)+5]
		}
	}
}

// Reset the internal State of the Curl sponge by filling it with all 0's.
func (c *Curl) Reset() {
	for i := range c.State {
		c.State[i] = 0
	}
}

// HashTrits returns the hash of the given trits.
func HashTrits(trits Trits, rounds ...CurlRounds) (Trits, error) {
	c := NewCurl(rounds...)
	c.Absorb(trits)
	return c.Squeeze(HashTrinarySize)
}

// HashTrytes returns the hash of the given trytes.
func HashTrytes(t Trytes, rounds ...CurlRounds) (Trytes, error) {
	c := NewCurl(rounds...)
	err := c.AbsorbTrytes(t)
	if err != nil {
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
	clone := NewCurl(c.Rounds).(*Curl)
	copy(clone.State, c.State)
	return clone
}
