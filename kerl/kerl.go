package kerl

import (
	"github.com/pkg/errors"
	"hash"

	. "github.com/iotaledger/iota.go/consts"
	keccak "github.com/iotaledger/iota.go/kerl/sha3"
	. "github.com/iotaledger/iota.go/trinary"
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

// Squeeze out length trits. Length has to be a multiple of HashTrinarySize.
func (k *Kerl) Squeeze(length int) (Trits, error) {
	if length%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	out := make(Trits, length)
	for i := 1; i <= length/HashTrinarySize; i++ {
		h := k.s.Sum(nil)
		ts, err := BytesToTrits(h)
		if err != nil {
			return nil, err
		}
		//ts[HashSize-1] = 0
		copy(out[HashTrinarySize*(i-1):HashTrinarySize*i], ts)
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
// This is only defined for Trit slices that are a multiple of HashTrinarySize long.
func (k *Kerl) Absorb(in Trits) error {
	if len(in) == 0 || len(in)%HashTrinarySize != 0 {
		return errors.Wrap(ErrInvalidTritsLength, "trits slice length must be a multiple of 243")
	}

	for i := 1; i <= len(in)/HashTrinarySize; i++ {
		// in[(HashSize*i)-1] = 0
		b, err := TritsToBytes(in[HashTrinarySize*(i-1) : HashTrinarySize*i])
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
