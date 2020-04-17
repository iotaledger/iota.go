// Package kerl implements the Kerl hashing function.
package kerl

import (
	"hash"
	"strings"

	. "github.com/iotaledger/iota.go/consts"
	keccak "github.com/iotaledger/iota.go/kerl/sha3"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

// Kerl is a to trinary aligned version of keccak
type Kerl struct {
	s hash.Hash
}

// NewKerl returns a new Kerl
func NewKerl() SpongeFunction {
	k := &Kerl{
		s: keccak.NewLegacyKeccak384(),
	}
	return k
}

func (k *Kerl) absorbBytes(in []byte) (err error) {
	_, err = k.s.Write(in)
	return
}

func (k *Kerl) squeezeBytes() ([]byte, error) {
	out := make([]byte, HashBytesSize)
	h := k.s.Sum(nil)

	// copy into out and fix the last trit
	copy(out, h)
	KerlBytesZeroLastTrit(out)

	// re-initialize keccak for the next squeeze
	k.Reset()
	for i := range h {
		h[i] = ^h[i]
	}
	if err := k.absorbBytes(h); err != nil {
		return nil, err
	}
	return out, nil
}

// Absorb fills the internal state of the sponge with the given trits.
// This is only defined for Trit slices that are a multiple of HashTrinarySize long.
func (k *Kerl) Absorb(in Trits) error {
	if len(in) == 0 || len(in)%HashTrinarySize != 0 {
		return errors.Wrap(ErrInvalidTritsLength, "trits slice length must be a multiple of 243")
	}

	for i := 0; i < len(in); i += HashTrinarySize {
		bs, err := KerlTritsToBytes(in[i : i+HashTrinarySize])
		if err != nil {
			return err
		}
		if err = k.absorbBytes(bs); err != nil {
			return err
		}
	}
	return nil
}

// AbsorbTrytes fills the internal State of the sponge with the given trytes.
func (k *Kerl) AbsorbTrytes(in Trytes) error {
	if len(in) == 0 || len(in)%HashTrytesSize != 0 {
		return errors.Wrap(ErrInvalidTrytesLength, "trytes length must be a multiple of 81")
	}

	for i := 0; i < len(in); i += HashTrytesSize {
		bs, err := KerlTrytesToBytes(in[i : i+HashTrytesSize])
		if err != nil {
			return err
		}
		if err = k.absorbBytes(bs); err != nil {
			return err
		}
	}
	return nil
}

// MustAbsorbTrytes fills the internal State of the sponge with the given trytes.
// It panics if the given trytes are not valid.
func (k *Kerl) MustAbsorbTrytes(inn Trytes) {
	err := k.AbsorbTrytes(inn)
	if err != nil {
		panic(err)
	}
}

// Squeeze out length trits. Length has to be a multiple of HashTrinarySize.
func (k *Kerl) Squeeze(length int) (Trits, error) {
	if length%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	out := make(Trits, length)
	for i := 0; i < length; i += HashTrinarySize {
		bs, err := k.squeezeBytes()
		if err != nil {
			return nil, err
		}
		ts, err := KerlBytesToTrits(bs)
		if err != nil {
			return nil, err
		}
		copy(out[i:], ts)
	}

	return out, nil
}

// MustSqueeze squeezes out trits of the given length. Length has to be a multiple of HashTrinarySize.
// It panics if the length is not valid.
func (k *Kerl) MustSqueeze(length int) Trits {
	out, err := k.Squeeze(length)
	if err != nil {
		panic(err)
	}
	return out
}

// SqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
func (k *Kerl) SqueezeTrytes(length int) (Trytes, error) {
	if length%HashTrinarySize != 0 {
		return "", ErrInvalidSqueezeLength
	}

	var out strings.Builder
	out.Grow(length / TritsPerTryte)

	for i := 0; i < length/HashTrinarySize; i++ {
		bs, err := k.squeezeBytes()
		if err != nil {
			return "", err
		}
		ts, err := KerlBytesToTrytes(bs)
		if err != nil {
			return "", err
		}
		out.WriteString(ts)
	}
	return out.String(), nil
}

// MustSqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
// It panics if the trytes or the length are not valid.
func (k *Kerl) MustSqueezeTrytes(length int) Trytes {
	out, err := k.SqueezeTrytes(length)
	if err != nil {
		panic(err)
	}
	return out
}

// Reset the internal state of the Kerl sponge.
func (k *Kerl) Reset() {
	k.s.Reset()
}

// Clone returns a deep copy of the current Kerl
func (k *Kerl) Clone() SpongeFunction {
	clone := NewKerl().(*Kerl)

	clone.s = keccak.CloneState(k.s)
	return clone
}
