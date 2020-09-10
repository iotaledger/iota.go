// Package kerl implements the Kerl hash function.
package kerl

import (
	"hash"
	"strings"
	"unsafe"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/kerl/sha3"
	. "github.com/iotaledger/iota.go/legacy/signing/utils"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/pkg/errors"
)

// ErrAbsorbAfterSqueeze is returned when absorb is called on the same hash after a squeeze.
var ErrAbsorbAfterSqueeze = errors.New("absorb after squeeze")

// kerlDirection indicates the direction bytes are flowing through the sponge.
type kerlDirection int

const (
	// kerlAbsorbing indicates that the sponge is absorbing input.
	kerlAbsorbing kerlDirection = iota
	// kerlSqueezing indicates that the sponge is being squeezed.
	kerlSqueezing
)

// Kerl is a to trinary aligned version of keccak
type Kerl struct {
	hash.Hash                            // underlying binary hashing function
	state     kerlDirection              // whether the sponge is absorbing or squeezing
	buf       [legacy.HashBytesSize]byte // internal buffer
}

// NewKerl returns a new Kerl
func NewKerl() *Kerl {
	return &Kerl{Hash: sha3.NewLegacyKeccak384(), state: kerlAbsorbing}
}

// notUnaligned flips each bit of the internal buffer.
func (k *Kerl) notUnaligned() {
	bw := (*[legacy.HashBytesSize / 8]uint64)(unsafe.Pointer(&k.buf))[: legacy.HashBytesSize/8 : legacy.HashBytesSize/8]
	bw[0] = ^bw[0]
	bw[1] = ^bw[1]
	bw[2] = ^bw[2]
	bw[3] = ^bw[3]
	bw[4] = ^bw[4]
	bw[5] = ^bw[5]
}

// squeezeSum squeezes the current hash sum into the hash's state.
func (k *Kerl) squeezeSum() {
	// absorb the new state, when squeezing more than once
	if k.state == kerlSqueezing {
		k.notUnaligned()
		k.Hash.Reset()
		k.Hash.Write(k.buf[:])
	}
	k.state = kerlSqueezing
	k.Hash.Sum(k.buf[:0])
}

// Write absorbs more data into the hash's state.
// In oder to have consistent behavior with Absorb and AbsorbTrytes, it must be assured that the bytes written are
// multiples of HashByteSize and do not represent ternary numbers with a non-zero 243rd trit, e.g. by calling
// KerlBytesZeroLastTrit or by only using output from the Kerl hash function.
func (k *Kerl) Write(in []byte) (int, error) {
	if k.state != kerlAbsorbing {
		return 0, ErrAbsorbAfterSqueeze
	}
	return k.Hash.Write(in)
}

// Read squeezes an arbitrary number of bytes. The buffer will be filled in multiples of HashByteSize.
func (k *Kerl) Read(b []byte) (n int, err error) {
	for len(b) >= legacy.HashBytesSize {
		k.squeezeSum()

		copy(b, k.buf[:])
		KerlBytesZeroLastTrit(b[:legacy.HashBytesSize])
		b = b[legacy.HashBytesSize:]
		n += legacy.HashBytesSize
	}
	return n, nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (k *Kerl) Sum(b []byte) []byte {
	// make a copy of k so that state and buffer are preserved
	dup := *k
	dup.squeezeSum()
	return append(b, dup.buf[:]...)
}

// Reset resets the Hash to its initial state.
func (k *Kerl) Reset() {
	k.Hash.Reset()
	k.state = kerlAbsorbing
}

// Size returns the number of bytes Sum will return.
func (k *Kerl) Size() int {
	return legacy.HashBytesSize
}

// Absorb fills the internal state of the sponge with the given trits.
// This is only defined for Trit slices that are a multiple of HashTrinarySize long.
func (k *Kerl) Absorb(in trinary.Trits) error {
	if len(in) == 0 || len(in)%legacy.HashTrinarySize != 0 {
		return errors.Wrap(legacy.ErrInvalidTritsLength, "trits slice length must be a multiple of 243")
	}

	// absorb all the chunks
	for len(in) >= legacy.HashTrinarySize {
		bs, err := KerlTritsToBytes(in[:legacy.HashTrinarySize])
		if err != nil {
			return err
		}
		if _, err = k.Write(bs); err != nil {
			return err
		}
		in = in[legacy.HashTrinarySize:]
	}
	return nil
}

// AbsorbTrytes fills the internal State of the sponge with the given trytes.
func (k *Kerl) AbsorbTrytes(in trinary.Trytes) error {
	if len(in) == 0 || len(in)%legacy.HashTrytesSize != 0 {
		return errors.Wrap(legacy.ErrInvalidTrytesLength, "trytes length must be a multiple of 81")
	}

	// absorb all the chunks
	for len(in) >= legacy.HashTrytesSize {
		bs, err := KerlTrytesToBytes(in[:legacy.HashTrytesSize])
		if err != nil {
			return err
		}
		if _, err = k.Write(bs); err != nil {
			return err
		}
		in = in[legacy.HashTrytesSize:]
	}
	return nil
}

// MustAbsorbTrytes fills the internal State of the sponge with the given trytes.
// It panics if the given trytes are not valid.
func (k *Kerl) MustAbsorbTrytes(in trinary.Trytes) {
	err := k.AbsorbTrytes(in)
	if err != nil {
		panic(err)
	}
}

// Squeeze out length trits. Length has to be a multiple of HashTrinarySize.
func (k *Kerl) Squeeze(length int) (trinary.Trits, error) {
	if length%legacy.HashTrinarySize != 0 {
		return nil, legacy.ErrInvalidSqueezeLength
	}

	out := make(trinary.Trits, 0, length)
	for i := 0; i < length/legacy.HashTrinarySize; i++ {
		k.squeezeSum()
		ts, err := KerlBytesToTrits(k.buf[:])
		if err != nil {
			return nil, err
		}
		out = append(out, ts...)
	}
	return out, nil
}

// MustSqueeze squeezes out trits of the given length. Length has to be a multiple of HashTrinarySize.
// It panics if the length is not valid.
func (k *Kerl) MustSqueeze(length int) trinary.Trits {
	out, err := k.Squeeze(length)
	if err != nil {
		panic(err)
	}
	return out
}

// SqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
func (k *Kerl) SqueezeTrytes(length int) (trinary.Trytes, error) {
	if length%legacy.HashTrinarySize != 0 {
		return "", legacy.ErrInvalidSqueezeLength
	}

	var out strings.Builder
	out.Grow(length / legacy.TritsPerTryte)

	for i := 0; i < length/legacy.HashTrinarySize; i++ {
		k.squeezeSum()
		ts, err := KerlBytesToTrytes(k.buf[:])
		if err != nil {
			return "", err
		}
		out.WriteString(ts)
	}
	return out.String(), nil
}

// MustSqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
// It panics if the trytes or the length are not valid.
func (k *Kerl) MustSqueezeTrytes(length int) trinary.Trytes {
	out, err := k.SqueezeTrytes(length)
	if err != nil {
		panic(err)
	}
	return out
}

// Clone returns a deep copy of the current Kerl
func (k *Kerl) Clone() SpongeFunction {
	clone := *k
	clone.Hash = sha3.CloneState(k.Hash)
	return &clone
}
