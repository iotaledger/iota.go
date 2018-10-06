package kerl

import (
	"fmt"
	"hash"

	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"
	keccak "golang.org/x/crypto/sha3"
)

var (
	ErrInvalidInputLength      = fmt.Errorf("output lengths must be of %d", TritHashLength)
	ErrInvalidInputTritsLength = fmt.Errorf("input trit slice must be a multiple of %d", TritHashLength)
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

// Squeeze out `length` trits. Length has to be a multiple of TritHashLength.
func (k *Kerl) Squeeze(length int) (Trits, error) {
	if length%curl.HashSize != 0 {
		return nil, ErrInvalidInputLength
	}

	out := make(Trits, length)
	for i := 1; i <= length/curl.HashSize; i++ {
		h := k.s.Sum(nil)
		ts, err := BytesToTrits(h)
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
// This is only defined for Trit slices that are a multiple of TritHashLength long.
func (k *Kerl) Absorb(in Trits) error {
	if len(in)%curl.HashSize != 0 {
		return ErrInvalidInputTritsLength
	}

	for i := 1; i <= len(in)/curl.HashSize; i++ {
		// in[(HashSize*i)-1] = 0
		b, err := TritsToBytes(in[curl.HashSize*(i-1) : curl.HashSize*i])
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
