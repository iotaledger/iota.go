// Package curl implements the Curl hashing function.
package curl

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
)

const (
	// StateSize is the size of the Curl hash function.
	StateSize = HashTrinarySize * 3
	// NumberOfRounds is the default number of rounds in transform.
	NumberOfRounds = 81
)

var (
	// optional transform function in C.
	transformC func(Trits)
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
	State Trits
}

// NewCurl initializes a new instance with an empty State.
func NewCurl() *Curl {
	c := &Curl{
		State: make(Trits, StateSize),
	}
	return c
}

// Squeeze out length trits. Length has to be a multiple of HashTrinarySize.
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

// Squeeze out length trits. Length has to be a multiple of HashTrinarySize.
// Returns trytes. It panics if the trytes are not valid.
func (c *Curl) SqueezeTrytes(length int) Trytes {
	trits, _ := c.Squeeze(length)
	return MustTritsToTrytes(trits)
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

// Absorb fills the internal State of the sponge with the given trytes.
// It panics if the given trytes are not valid.
func (c *Curl) AbsorbTrytes(inn Trytes) {
	var in Trits
	if len(inn) == 0 {
		in = Trits{0}
	} else {
		in = MustTrytesToTrits(inn)
	}
	c.Absorb(in)
}

// Transform does Transform in sponge func.
func (c *Curl) Transform() {
	if transformC != nil {
		transformC(c.State)
		return
	}

	var cpy [StateSize]int8

	for r := NumberOfRounds; r > 0; r-- {
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

// HashTrits returns hash trits of the given trits.
func HashTrits(trits Trits) (Trits, error) {
	c := NewCurl()
	c.Absorb(trits)
	return c.Squeeze(HashTrinarySize)
}

// HashTrytes returns hash trytes of the given trytes.
// It panics if the given trytes are not valid.
func HashTrytes(t Trytes) Trytes {
	c := NewCurl()
	c.AbsorbTrytes(t)
	return c.SqueezeTrytes(HashTrinarySize)
}
