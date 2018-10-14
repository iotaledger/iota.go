package curl

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
)

const (
	StateSize      = HashTrinarySize * 3
	NumberOfRounds = 81
)

var (
	transformC func(Trits)
	TruthTable = [11]int8{1, 0, -1, 2, 1, -1, 0, 2, -1, 1, 0}
	Indices    [StateSize + 1]int
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

//Squeeze do Squeeze in sponge func.
func (c *Curl) Squeeze() Trytes {
	ret := MustTritsToTrytes(c.State[:HashTrinarySize])
	c.Transform()

	return ret
}

// Absorb fills the internal State of the sponge with the given trits.
// It panics if the given trytes are not valid.
func (c *Curl) Absorb(inn Trytes) {
	var in Trits
	if len(inn) == 0 {
		in = Trits{0}
	} else {
		in = MustTrytesToTrits(inn)
	}
	var lenn int
	for i := 0; i < len(in); i += lenn {
		lenn = HashTrinarySize

		if len(in)-i < HashTrinarySize {
			lenn = len(in) - i
		}

		copy(c.State, in[i:i+lenn])
		c.Transform()
	}
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
		c.State = c.State[:StateSize]
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

// HashTrytes returns hash of t.
func HashTrits(trits Trits) Trits {
	c := NewCurl()
	c.Absorb(MustTritsToTrytes(trits))
	return MustTrytesToTrits(c.Squeeze())
}

// HashTrytes returns hash of t.
func HashTrytes(t Trytes) Trytes {
	c := NewCurl()
	c.Absorb(t)
	return c.Squeeze()
}
