// Package kerl implements the Kerl hashing function.
package kerl

import (
	"hash"
	"unsafe"

	"github.com/pkg/errors"

	. "github.com/iotaledger/iota.go/consts"
	keccak "github.com/iotaledger/iota.go/kerl/sha3"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
)

// Kerl is a to trinary aligned version of keccak
type Kerl struct {
	s hash.Hash
}

// NewKerl returns a new Kerl
func NewKerl() SpongeFunction {
	k := &Kerl{
		s: keccak.New384(),
	}
	return k
}

func (k *Kerl) squeezeBytes(numChunks int) ([]byte, error) {
	result := make([]byte, numChunks*HashBytesSize)
	buffer := make([]byte, HashBytesSize)

	for i := 1; i <= numChunks; i++ {
		h := k.s.Sum(buffer[:0])

		// copy into result and fix the last trit
		chunk := result[HashBytesSize*(i-1) : HashBytesSize*i]
		copy(chunk, h)
		KerlBytesZeroLastTrit(chunk)

		// re-initialize keccak for the next squeeze
		k.s.Reset()
		for i, e := range h {
			h[i] = ^e
		}
		if _, err := k.s.Write(h); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Squeeze out length trits. Length has to be a multiple of HashTrinarySize.
func (k *Kerl) Squeeze(length int) (Trits, error) {
	if length%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	bs, err := k.squeezeBytes(length / HashTrinarySize)
	if err != nil {
		return nil, err
	}

	out := make(Trits, length)
	for i := 1; i <= length/HashTrinarySize; i++ {
		ts, err := KerlBytesToTrits(bs[HashBytesSize*(i-1) : HashBytesSize*i])
		if err != nil {
			return nil, err
		}
		copy(out[HashTrinarySize*(i-1):HashTrinarySize*i], ts)
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

	bs, err := k.squeezeBytes(length / HashTrinarySize)
	if err != nil {
		return "", err
	}

	out := make([]byte, length/HashTrinarySize*HashTrytesSize)
	for i := 1; i <= length/HashTrinarySize; i++ {
		ts, err := KerlBytesToTrytes(bs[HashBytesSize*(i-1) : HashBytesSize*i])
		if err != nil {
			return "", err
		}
		copy(out[HashTrytesSize*(i-1):HashTrytesSize*i], ts)
	}

	// convert into trytes without copying
	return *(*string)(unsafe.Pointer(&out)), nil
}

// MustSqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
// It panics if the trytes or the length are not valid.
func (k *Kerl) MustSqueezeTrytes(length int) Trytes {
	return MustTritsToTrytes(k.MustSqueeze(length))
}

// Absorb fills the internal state of the sponge with the given trits.
// This is only defined for Trit slices that are a multiple of HashTrinarySize long.
func (k *Kerl) Absorb(in Trits) error {
	if len(in) == 0 || len(in)%HashTrinarySize != 0 {
		return errors.Wrap(ErrInvalidTritsLength, "trits slice length must be a multiple of 243")
	}

	for i := 1; i <= len(in)/HashTrinarySize; i++ {
		b, err := KerlTritsToBytes(in[HashTrinarySize*(i-1) : HashTrinarySize*i])
		if err != nil {
			return err
		}
		if _, err := k.s.Write(b); err != nil {
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

	for i := 1; i <= len(in)/HashTrytesSize; i++ {
		b, err := KerlTrytesToBytes(in[HashTrytesSize*(i-1) : HashTrytesSize*i])
		if err != nil {
			return err
		}
		if _, err := k.s.Write(b); err != nil {
			return err
		}
	}
	return nil
}

// AbsorbTrytes fills the internal State of the sponge with the given trytes.
// It panics if the given trytes are not valid.
func (k *Kerl) MustAbsorbTrytes(inn Trytes) {
	err := k.AbsorbTrytes(inn)
	if err != nil {
		panic(err)
	}
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
