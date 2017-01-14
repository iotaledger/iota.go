package giota

import ()

const (
	HashSize         = 243
	AddressTritsLen  = HashSize
	AddressTrytesLen = AddressTritsLen / NumberOfTritsPerTryte
	StateSize        = 729
)

var (
	TruthTable = []int{1, 0, -1, 0, 1, -1, 0, 0, -1, 1, 0}
	Indices    = make([]int, StateSize+1)
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
	state     []int
	stateCopy []int
}

// NewCurl initializes a new instance with an empty state, which
// is equivalent to
// 		c := &Curl{}
// 		c.Init([]int{})
func NewCurl() *Curl {
	c := &Curl{}
	c.Init([]int{})

	return c
}

func (c *Curl) Init(state []int) {
	c.state = make([]int, StateSize)
	c.stateCopy = make([]int, StateSize)

	if state != nil {
		copy(c.state, state)
	}
}

func (c *Curl) Squeeze() []int {
	ret := make([]int, HashSize)
	copy(ret, c.state[:HashSize])
	c.Transform()

	return ret
}

// Absorb fills the internal state of the sponge with the given trits.
func (c *Curl) Absorb(in []int) {
	c.absorbWithOffset(in, 0, len(in))
}

func (c *Curl) absorbWithOffset(in []int, offset, size int) {
	for {
		len := 243
		if size < 243 {
			len = size
		}
		copy(c.state[0:len], in[offset:offset+len])
		c.Transform()
		offset += 243
		size -= 243
		if size <= 0 {
			break
		}
	}
}

// Transforms
func (c *Curl) Transform() {
	for r := 27; r > 0; r-- {
		copy(c.stateCopy, c.state)
		for i := 0; i < StateSize; {
			c.state[i] = TruthTable[c.stateCopy[Indices[i]]+(c.stateCopy[Indices[i+1]]<<2)+5]
			i++
		}
	}
}

// State returns the internal state slice.
func (c *Curl) State() []int {
	return c.state
}

// Reset the internal state of the Curl sponge by filling it with all
// 0's.
func (c *Curl) Reset() {
	for i := range c.state {
		c.state[i] = 0
	}
	copy(c.stateCopy, c.state)
}
