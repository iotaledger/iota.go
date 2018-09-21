package curl

import (
	"github.com/iotaledger/giota/trinary"
)

// constants for Sizes.
const (
	HashSize       = 243
	HashSizeTrytes = 81
	StateSize      = HashSize * 3
	NumberOfRounds = 81
)

var (
	// EmptyHash represents an empty hash.
	EmptyHash trinary.Trytes = "999999999999999999999999999999999999999999999999999999999999999999999999999999999"
)

var (
	transformC func(trinary.Trits)
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
	State trinary.Trits
}

// NewCurl initializes a new instance with an empty State.
func NewCurl() *Curl {
	c := &Curl{
		State: make(trinary.Trits, StateSize),
	}
	return c
}

//Squeeze do Squeeze in sponge func.
func (c *Curl) Squeeze() trinary.Trytes {
	ret := c.State[:HashSize].Trytes()
	c.Transform()

	return ret
}

// Absorb fills the internal State of the sponge with the given trits.
func (c *Curl) Absorb(inn trinary.Trytes) {
	in := inn.Trits()
	var lenn int
	for i := 0; i < len(in); i += lenn {
		lenn = trinary.TritHashLength

		if len(in)-i < trinary.TritHashLength {
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

// Hash returns hash of t.
func Hash(t trinary.Trytes) trinary.Trytes {
	c := NewCurl()
	c.Absorb(t)
	return c.Squeeze()
}
