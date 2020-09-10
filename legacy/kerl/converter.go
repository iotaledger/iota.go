// Package kerl implements the Kerl hashing function.
package kerl

import (
	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/kerl/bigint"
	. "github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/pkg/errors"
)

const (
	// the middle of the domain described by one tryte
	halfTryte = legacy.TryteRadix / 2

	// largest number of trytes that can be represented as a uint32
	trytesPerUint32 = 6
	// radix used in the uint32 chunk conversion, i.e. 27^trytesPerUint32
	uint32Radix = 387420489
	// number of uint32 chunks to represent the hash, i.e. ceil(HashTrytesSize / trytesPerUint32)
	hashUint32Size = (legacy.HashTrytesSize + trytesPerUint32 - 1) / trytesPerUint32
)

var (
	// maxTer243 represents the largest value that can be represented in 243-trit balanced ternary, i.e. ⌊3²⁴³ / 2⌋.
	// since ⌊3²⁴³ / 2⌋ cannot be expressed in 384 bits, it only consists of its lower 384 bits.
	maxTer243 = bigint.MustParseU384("0x1b3dc3cef97f039efe13f810fde381a1da330aa36cee5506f1c6d805246cd94ab09a028b3d8cf0eedd01633cf16b9c2d")

	// maxTer242 represents the largest value that can be represented in 242-trit balanced ternary, i.e. ⌊3²⁴² / 2⌋.
	maxTer242 = bigint.MustParseU384("0x5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964")
	// maxTer242 represents the smallest value that can be represented in 242-trit balanced ternary, i.e. -⌊3²⁴² / 2⌋.
	minTer242 = bigint.MustParseU384("0xa19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c")
	// trit243 represents the value of the 243rd trit, i.e. 3²⁴².
	trit243 = bigint.MustParseU384("0xbcd3d7df50ff57bf540d500b53ed011691775c6cf3498e04a12f3aae184890dc75bc01b22908a09f3e00ecd34b9d12c9")
)

func tryteValuesToTrits(vs []int8) Trits {
	trits := make([]int8, len(vs)*legacy.TritsPerTryte)
	for i, v := range vs {
		MustPutTryteTrits(trits[i*legacy.TritsPerTryte:], v)
	}
	return trits
}

func tryteValuesToTrytes(vs []int8) Trytes {
	if len(vs) != legacy.HashTrytesSize {
		panic(legacy.ErrInvalidTrytesLength) // bounds check hint to compiler
	}
	trytes := make([]byte, legacy.HashTrytesSize)
	for i := range vs {
		trytes[i] = MustTryteValueToTryte(vs[i])
	}
	return string(trytes)
}

// tryteValueZeroLastTrit returns the value of tryte v, with the last trit set to zero.
// It takes a tryte value of three trits a+3b+9c and returns a+3b.
func tryteValueZeroLastTrit(v int8) int8 {
	if v > 4 {
		return v - 9
	}
	if v < -4 {
		return v + 9
	}
	return v
}

func tryteValuesToBytes(vs []int8) []byte {
	if len(vs) != legacy.HashTrytesSize { // hint to the compiler that vs has constant length
		panic(legacy.ErrInvalidTrytesLength)
	}

	// assure that the last trit is zero
	vs[legacy.HashTrytesSize-1] = tryteValueZeroLastTrit(vs[legacy.HashTrytesSize-1])

	// convert the balanced ternary input to base 27⁶, the largest power of 27 still fitting into an uint32
	// add ⌊27 / 2⌋ to each tryte value to get base 27
	cs := make([]uint32, hashUint32Size)
	for i := range cs {
		for j := trytesPerUint32 - 1; j >= 0; j-- {
			idx := uint(i*trytesPerUint32 + j) // hint to the compiler that idx is always non-negative
			if idx < legacy.HashTrytesSize {
				cs[i] = legacy.TryteRadix*cs[i] + uint32(vs[idx]+halfTryte)
			}
		}
	}

	// convert the base 27⁶ number to a 384-bit unsigned integer
	b := bigint.U384()
	bs := b.Words()[:0] // do not modify b directly, but work on a new slice backed by the same array
	for i := len(cs) - 1; i >= 0; i-- {
		n := len(bs)
		// multiply by the radix and add the value of the next chunk
		carry := cs[i]
		for i := 0; i < n; i++ {
			v := uint32Radix*uint64(bs[i]) + uint64(carry)
			carry, bs[i] = uint32(v>>32), uint32(v)
		}
		// increase length of slice if necessary
		if carry > 0 && n < cap(bs) {
			bs = append(bs, carry)
		}
	}

	// since we initially added ⌊27 / 2⌋ to each of the 81 tryte values, we now need to subtract ⌊27⁸¹ / 2⌋ = ⌊3²⁴³ / 2⌋
	// this leads the correct two's complement representation for negative numbers
	b.Sub(maxTer243)

	// return the corresponding byte slice
	bytes := make([]byte, legacy.HashBytesSize)
	b.Read(bytes)
	return bytes
}

// KerlBytesZeroLastTrit changes a chunk of 48 bytes so that the corresponding ternary number has 242th trit set to 0.
func KerlBytesZeroLastTrit(bytes []byte) {
	// bytes represents a signed 384-bit integer
	b := bigint.U384()
	b.SetBytes(bytes)

	// assure that the value is in [-⌊3²⁴² / 2⌋, ⌊3²⁴² / 2⌋]
	if b.MSB() != 0 {
		if b.Cmp(minTer242) >= 0 {
			return
		}
		// add 3²⁴² if the last trit was -1, i.e. if the value is less than -⌊3²⁴² / 2⌋
		b.Add(trit243)
	} else {
		if b.Cmp(maxTer242) <= 0 {
			return
		}
		// subtract 3²⁴² if the last trit was 1, i.e. if the value is greater than ⌊3²⁴² / 2⌋
		b.Sub(trit243)
	}

	// update the bytes if we made changes
	b.Read(bytes)
}

// KerlTritsToBytes is only defined for hashes, i.e. chunks of trits of length 243. It returns 48 bytes.
func KerlTritsToBytes(trits Trits) ([]byte, error) {
	if !CanBeHash(trits) {
		return nil, errors.Wrapf(legacy.ErrInvalidTritsLength, "must be %d in size", legacy.HashTrinarySize)
	}

	// convert to tryte values
	vs := make([]int8, legacy.HashTrytesSize)
	for i := range vs {
		tryteTrits := trits[i*legacy.TritsPerTryte:]
		_ = tryteTrits[2] // bounds check hint to compiler
		vs[i] = tryteTrits[0] + tryteTrits[1]*3 + tryteTrits[2]*9
	}

	return tryteValuesToBytes(vs), nil
}

// KerlTrytesToBytes is only defined for hashes, i.e. chunks of trytes of length 81. It returns 48 bytes.
func KerlTrytesToBytes(trytes Trytes) ([]byte, error) {
	if len(trytes) != legacy.HashTrytesSize {
		return nil, errors.Wrapf(legacy.ErrInvalidTrytesLength, "must be %d in size", legacy.HashBytesSize)
	}

	// convert to tryte values
	vs := make([]int8, legacy.HashTrytesSize)
	for i := 0; i < legacy.HashTrytesSize; i++ {
		vs[i] = MustTryteToTryteValue(trytes[i])
	}

	return tryteValuesToBytes(vs), nil
}

// KerlBytesToTrits is only defined for hashes, i.e. chunks of 48 bytes. It returns 243 trits.
func KerlBytesToTrits(b []byte) (Trits, error) {
	if len(b) != legacy.HashBytesSize {
		return nil, errors.Wrapf(legacy.ErrInvalidBytesLength, "must be %d in size", legacy.HashBytesSize)
	}

	vs := make([]int8, legacy.HashTrytesSize)
	bytesToTryteValues(b, vs)
	return tryteValuesToTrits(vs), nil
}

// KerlBytesToTrytes is only defined for hashes, i.e. chunks of 48 bytes. It returns 81 trytes.
func KerlBytesToTrytes(b []byte) (Trytes, error) {
	if len(b) != legacy.HashBytesSize {
		return "", errors.Wrapf(legacy.ErrInvalidBytesLength, "must be %d in size", legacy.HashBytesSize)
	}

	vs := make([]int8, legacy.HashTrytesSize)
	bytesToTryteValues(b, vs)
	return tryteValuesToTrytes(vs), nil
}
