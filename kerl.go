/*
MIT License

Copyright (c) 2017 Sascha Hanse

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

import (
	"fmt"
	"hash"

	keccak "github.com/tildeleb/hashland/keccakpg"
)

// Kerl ...
// TODO: find out the difference between this anc Curl and document
type Kerl struct {
	s hash.Hash
}

// NewKerl returns a new Kerl
func NewKerl() *Kerl {
	k := &Kerl{
		s: keccak.New384(),
	}
	return k
}

// Squeeze out `length` trits. Length has to be a multiple of TritHashLength.
func (k *Kerl) Squeeze(length int) (Trits, error) {
	if length%HashSize != 0 {
		return nil, fmt.Errorf("Squeeze is only defined for output lengths slices that are a multiple of %d",
			TritHashLength)
	}

	out := make(Trits, length)
	for i := 1; i <= length/HashSize; i++ {
		h := k.s.Sum(nil)
		ts, err := BytesToTrits(h)
		if err != nil {
			return nil, err
		}
		//ts[HashSize-1] = 0
		copy(out[HashSize*(i-1):HashSize*i], ts)
		k.s.Reset()
		for i, e := range h {
			h[i] = ^e
		}
		k.s.Write(h)
	}

	return out, nil
}

// Absorb fills the internal state of the sponge with the given trits.
// This is only defined for Trit slices that are a multiple of TritHashLength long.
func (k *Kerl) Absorb(in Trits) error {
	if len(in)%HashSize != 0 {
		return fmt.Errorf("Absorb is only defined for Trit slices that are a multiple of %d long", TritHashLength)
	}

	for i := 1; i <= len(in)/HashSize; i++ {
		// in[(HashSize*i)-1] = 0
		b, err := in[HashSize*(i-1) : HashSize*i].Bytes()
		if err != nil {
			return err
		}
		k.s.Write(b)
	}

	return nil
}

// Reset the internal state of the Kerl sponge.
func (k *Kerl) Reset() {
	k.s.Reset()
}
