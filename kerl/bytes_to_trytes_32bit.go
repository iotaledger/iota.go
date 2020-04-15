// architectures that do not have 64bit division instructions
// +build arm 386 mips mipsle

// Package kerl implements the Kerl hashing function.
package kerl

import (
	"math"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
)

const (
	// largest number of trytes that can be represented as a uint16
	trytesPerUint16 = 3
	// radix used in the uint16 chunk conversion, i.e. 27^trytesPerUint16
	uint16Radix = 19683
	// number of uint16 chunks to represent the hash, i.e. ceil(HashTrytesSize / trytesPerUint16)
	hashUint16Size = (HashTrytesSize + trytesPerUint16 - 1) / trytesPerUint16
)

func uint16ToTryteValues(cs []uint16, vs []int8) {
	for i, c := range cs {
		tmp := vs[i*trytesPerUint16:]
		_ = tmp[2] // bounds check hint to compiler
		// unroll all the divisions
		c, tmp[0] = c/TryteRadix, int8(c%TryteRadix)-halfTryte
		c, tmp[1] = c/TryteRadix, int8(c%TryteRadix)-halfTryte
		tmp[2] = int8(c) - halfTryte
	}
}

func bytesToTryteValues(bytes []byte, vs []int8) {
	// bytes represents a signed 384-bit integer, which is always greater -⌊3²⁴³ / 2⌋ and less ⌊3²⁴³ / 2⌋
	b := bigint.U384()
	b.SetBytes(bytes)

	// whether bytes represents a negative number in two's complement
	negative := b.MSB() != 0

	// add ⌊3²⁴³ / 2⌋ and treat the result as an unsigned integer
	// since maxTer243 only contains the lower 384 bits of ⌊3²⁴³ / 2⌋, we need to consider the carry
	carry := b.Add(maxTer243)
	if !negative {
		// maxTer243 misses the leading bit of ⌊3²⁴³ / 2⌋ to fit into 384 bits
		// for negative numbers this cancels out, but for non-negative it needs to be considered
		carry += 1
	}

	// convert the 384-bit unsigned integer to base 27³, the largest power of 27 still fitting into an uint16
	cs := make([]uint16, hashUint16Size)
	bs := b.Words() // do not modify b directly, but work on a new slice backed by the same array
	rem := carry    // use carry as the initial remainder
	for i := range cs {
		n := len(bs)
		// divide the entire integer by the radix
		for i := n - 1; i >= 0; i-- {
			hi, low := bs[i]>>16, bs[i]&math.MaxUint16

			v := (rem << 16) | hi
			hi, rem = v/uint16Radix, v%uint16Radix
			v = (rem << 16) | low
			low, rem = v/uint16Radix, v%uint16Radix

			bs[i] = (hi << 16) | low
		}
		// decrease length of slice to ignore leading zeros
		if n > 0 && bs[n-1] == 0 {
			bs = bs[:n-1]
		}

		cs[i] = uint16(rem)
		rem = 0 // reset the remainder for the next chunk
	}

	// convert the base 27³ number to balanced ternary
	// since we initially added ⌊3²⁴³ / 2⌋ = ⌊27⁸¹ / 2⌋, we now need to sub ⌊27 / 2⌋ from each of the 81 tryte values
	uint16ToTryteValues(cs, vs)
	// set the last trit to zero
	vs[HashTrytesSize-1] = tryteValueZeroLastTrit(vs[HashTrytesSize-1])
}
