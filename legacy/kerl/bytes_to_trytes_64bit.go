// +build !arm,!386,!mips,!mipsle

// Package kerl implements the Kerl hashing function.
package kerl

import (
	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/kerl/bigint"
)

func uint32ToTryteValues(cs []uint32, vs []int8) {
	for i, c := range cs {
		tmp := vs[i*trytesPerUint32:]
		// as HashTrytesSize is not a multiple of trytesPerUint32, handle the last uint32 chunk differently
		if len(tmp) < trytesPerUint32 {
			_ = tmp[2] // bounds check hint to compiler
			c, tmp[0] = c/legacy.TryteRadix, int8(c%legacy.TryteRadix)-halfTryte
			c, tmp[1] = c/legacy.TryteRadix, int8(c%legacy.TryteRadix)-halfTryte
			tmp[2] = int8(c) - halfTryte
			return
		}
		// unroll all the divisions
		c, tmp[0] = c/legacy.TryteRadix, int8(c%legacy.TryteRadix)-halfTryte
		c, tmp[1] = c/legacy.TryteRadix, int8(c%legacy.TryteRadix)-halfTryte
		c, tmp[2] = c/legacy.TryteRadix, int8(c%legacy.TryteRadix)-halfTryte
		c, tmp[3] = c/legacy.TryteRadix, int8(c%legacy.TryteRadix)-halfTryte
		c, tmp[4] = c/legacy.TryteRadix, int8(c%legacy.TryteRadix)-halfTryte
		tmp[5] = int8(c) - halfTryte
	}
	panic("unreachable")
}

// bytesToTryteValues converts bytes into its corresponding 81-tryte representation vs.
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

	// convert the 384-bit unsigned integer to base 27⁶, the largest power of 27 still fitting into an uint32
	cs := make([]uint32, hashUint32Size)
	bs := b.Words() // do not modify b directly, but work on a new slice backed by the same array
	rem := carry    // use carry as the initial remainder
	for i := range cs {
		n := len(bs)
		// divide the entire integer by the radix
		for i := n - 1; i >= 0; i-- {
			v := (uint64(rem) << 32) | uint64(bs[i])
			bs[i], rem = uint32(v/uint32Radix), uint32(v%uint32Radix)
		}
		// decrease length of slice to ignore leading zeros
		if n > 0 && bs[n-1] == 0 {
			bs = bs[:n-1]
		}

		cs[i] = rem
		rem = 0 // reset the remainder for the next chunk
	}

	// convert the base 27⁶ number to balanced ternary
	// since we initially added ⌊3²⁴³ / 2⌋ = ⌊27⁸¹ / 2⌋, we now need to sub ⌊27 / 2⌋ from each of the 81 tryte values
	uint32ToTryteValues(cs, vs)
	// set the last trit to zero
	vs[legacy.HashTrytesSize-1] = tryteValueZeroLastTrit(vs[legacy.HashTrytesSize-1])
}
