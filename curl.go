/*
MIT License

Copyright (c) 2016 Sascha Hanse
Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package giota

//constants for Sizes.
const (
	stateSize = 729
)

var (
	truthTable = Trits{1, 0, -1, 0, 1, -1, 0, 0, -1, 1, 0}
	indices    [stateSize + 1]int
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
	state Trits
}

// NewCurl initializes a new instance with an empty state.
func NewCurl() *Curl {
	c := &Curl{
		state: make(Trits, stateSize),
	}
	return c
}

//Squeeze do Squeeze in sponge func.
func (c *Curl) Squeeze() Trits {
	ret := make(Trits, HashSize)
	copy(ret, c.state[:HashSize])
	c.Transform()

	return ret
}

// Absorb fills the internal state of the sponge with the given trits.
func (c *Curl) Absorb(in Trits) {
	var lenn int
	for i := 0; i < len(in); i += lenn {
		lenn = 243
		if len(in)-i < 243 {
			lenn = len(in) - i
		}
		copy(c.state, in[i:i+lenn])
		c.Transform()
	}
}

// Transform does Transform in sponge func.
func (c *Curl) Transform() {
	cpy := make(Trits, stateSize)
	for r := 27; r > 0; r-- {
		copy(cpy, c.state)
		for i := 0; i < stateSize; i++ {
			c.state[i] = truthTable[cpy[indices[i]]+(cpy[indices[i+1]]<<2)+5]
		}
	}
}

// Reset the internal state of the Curl sponge by filling it with all
// 0's.
func (c *Curl) Reset() {
	for i := range c.state {
		c.state[i] = 0
	}
}

//Hash returns hash of t.
func (t Trits) Hash() Trits {
	c := NewCurl()
	c.Absorb(t)
	return c.Squeeze()
}
