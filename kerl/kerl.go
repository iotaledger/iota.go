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

package kerl

import (
	"fmt"
	"hash"

	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/trinary"
	keccak "golang.org/x/crypto/sha3"
)

var (
	ErrInvalidInputLength      = fmt.Errorf("output lengths must be of %d", trinary.TritHashLength)
	ErrInvalidInputTritsLength = fmt.Errorf("input trit slice must be a multiple of %d", trinary.TritHashLength)
)

// Kerl is a to trinary aligned version of keccak
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

// Squeeze out `length` trits. Length has to be a multiple of trinary.TritHashLength.
func (k *Kerl) Squeeze(length int) (trinary.Trits, error) {
	if length%curl.HashSize != 0 {
		return nil, ErrInvalidInputLength
	}

	out := make(trinary.Trits, length)
	for i := 1; i <= length/curl.HashSize; i++ {
		h := k.s.Sum(nil)
		ts, err := trinary.BytesToTrits(h)
		if err != nil {
			return nil, err
		}
		//ts[HashSize-1] = 0
		copy(out[curl.HashSize*(i-1):curl.HashSize*i], ts)
		k.s.Reset()
		for i, e := range h {
			h[i] = ^e
		}
		if _, err := k.s.Write(h); err != nil {
			return nil, err
		}
	}

	return out, nil
}

// Absorb fills the internal state of the sponge with the given trits.
// This is only defined for Trit slices that are a multiple of trinary.TritHashLength long.
func (k *Kerl) Absorb(in trinary.Trits) error {
	if len(in)%curl.HashSize != 0 {
		return ErrInvalidInputTritsLength
	}

	for i := 1; i <= len(in)/curl.HashSize; i++ {
		// in[(HashSize*i)-1] = 0
		b, err := in[curl.HashSize*(i-1) : curl.HashSize*i].Bytes()
		if err != nil {
			return err
		}
		if _, err := k.s.Write(b); err != nil {
			return err
		}
	}

	return nil
}

// Reset the internal state of the Kerl sponge.
func (k *Kerl) Reset() {
	k.s.Reset()
}
