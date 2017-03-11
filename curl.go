package giota

//constants for Sizes.
const (
	stateSize = 729
)

var (
	truthTable = Trits{1, 0, -1, 0, 1, -1, 0, 0, -1, 1, 0}
	indices    = make([]int, stateSize+1)
)

func init() {
	for i := 0; i < stateSize; i++ {
		p := -365
		if indices[i] < 365 {
			p = 364
		}
		indices[i+1] = indices[i] + p
	}
}

// Curl is a sponge function with an internal state of size StateSize.
// b = r + c, b = StateSize, r = HashSize, c = StateSize - HashSize
type Curl struct {
	State Trits
}

// NewCurl initializes a new instance with an empty state, which
// is equivalent to
// 		c := &Curl{}
// 		c.Init([]int{})
func NewCurl() *Curl {
	c := &Curl{
		State: make(Trits, stateSize),
	}
	return c
}

//Squeeze do Squeeze in sponge func.
func (c *Curl) Squeeze() Trits {
	ret := make(Trits, HashSize)
	copy(ret, c.State[:HashSize])
	c.Transform()

	return ret
}

// Absorb fills the internal state of the sponge with the given trits.
func (c *Curl) Absorb(in Trits) {
	lenn := 0
	for i := 0; i < len(in); i += lenn {
		lenn = 243
		if len(in)-i < 243 {
			lenn = len(in) - i
		}
		copy(c.State, in[i:i+lenn])
		c.Transform()
	}
}

// Transform does Transform in sponge func.
func (c *Curl) Transform() {
	cpy := make(Trits, stateSize)
	for r := 27; r > 0; r-- {
		copy(cpy, c.State)
		for i := 0; i < stateSize; i++ {
			c.State[i] = truthTable[cpy[indices[i]]+(cpy[indices[i+1]]<<2)+5]
		}
	}
}

// Reset the internal state of the Curl sponge by filling it with all
// 0's.
func (c *Curl) Reset() {
	for i := range c.State {
		c.State[i] = 0
	}
}

//Hash returns hash of in.
func (t Trits) Hash() Trits {
	c := NewCurl()
	c.Absorb(t)
	return c.Squeeze()
}
