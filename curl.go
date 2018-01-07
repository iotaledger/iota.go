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

// constants for Sizes.
const (
	stateSize      = 729
	numberOfRounds = 81
)

var (
	transformC func(Trits)
	truthTable = [11]int8{1, 0, -1, 2, 1, -1, 0, 2, -1, 1, 0}
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
func (c *Curl) Squeeze() Trytes {
	ret := c.state[:HashSize].Trytes()
	c.Transform()

	return ret
}

// Absorb fills the internal state of the sponge with the given trits.
func (c *Curl) Absorb(inn Trytes) {
	in := inn.Trits()
	var lenn int
	for i := 0; i < len(in); i += lenn {
		lenn = TritHashLength

		if len(in)-i < TritHashLength {
			lenn = len(in) - i
		}

		copy(c.state, in[i:i+lenn])
		c.Transform()
	}
}

// Transform does Transform in sponge func.
func (c *Curl) Transform() {
	if transformC != nil {
		transformC(c.state)
		return
	}

	var cpy [stateSize]int8

	for r := numberOfRounds; r > 0; r-- {
		copy(cpy[:], c.state)
		c.state = c.state[:stateSize]
		for i := 0; i < stateSize; i++ {
			t1 := indices[i]
			t2 := indices[i+1]
			c.state[i] = truthTable[cpy[t1]+(cpy[t2]<<2)+5]
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

// Hash returns hash of t.
func (t Trytes) Hash() Trytes {
	c := NewCurl()
	c.Absorb(t)
	return c.Squeeze()
}
