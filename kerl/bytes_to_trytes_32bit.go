// +build 386 arm mips

// Package kerl implements the Kerl hashing function.
package kerl

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
)

func bytesToTryteValues(bytes []byte) []int8 {
	b := make([]uint32, IntLength)
	bigintPutBytes(b, bytes)

	// the two's complement representation is only correct, if the number fits
	// into 48 bytes, i.e. has the 243th trit set to 0
	bigintZeroLastTrit(b)

	// convert to the unsigned bigint representing non-balanced ternary
	bigint.MustAdd(b, halfThree)

	vs := make([]int8, HashTrytesSize)

	// initially, all words of the bigint are non-zero
	nzIndex := IntLength - 1
	for i := 0; i < HashTrytesSize-1; i++ {
		// divide the bigint by the radix
		var rem uint32
		for i := nzIndex; i >= 0; i-- {
			upper, lower := b[i]>>16, b[i]&0xFFFF

			v := (rem << 16) | upper
			upper = v / tryteRadix
			rem = v % tryteRadix

			v = (rem << 16) | lower
			lower = v / tryteRadix
			rem = v % tryteRadix

			b[i] = (upper << 16) | lower
		}
		// the tryte value is the remainder converted back to balanced ternary
		vs[i] = int8(rem) - halfTryte

		// decrement index, if the highest considered word of the bigint turned zero
		if nzIndex > 0 && b[nzIndex] == 0 {
			nzIndex--
		}
	}

	// special case for the last tryte, where no further division is necessary
	vs[HashTrytesSize-1] = tryteZeroLastTrit(int8(b[0]) - halfTryte)

	return vs
}
