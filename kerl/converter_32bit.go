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
			upper, rem = v/tryteRadix, v%tryteRadix
			v = (rem << 16) | lower
			lower, rem = v/tryteRadix, v%tryteRadix

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
	vs[HashTrytesSize-1] = tryteValueZeroLastTrit(int8(b[0]) - halfTryte)

	return vs
}

func tryteValuesToBytes(vs []int8) []byte {
	b := make([]uint32, IntLength)
	// set the last trit of the last tryte to zero
	v := tryteValueZeroLastTrit(vs[HashTrytesSize-1])
	// initialize the first part of the bigint with the non-balanced representation of this 2-trit value
	b[0] = uint32(v + 4)

	// initially, all words of the bigint are zero
	nzIndex := 0
	for i := HashTrytesSize - 2; i >= 0; i-- {
		// first, multiply the bigint by the radix
		var carry uint32
		for i := 0; i <= nzIndex; i++ {
			upper, lower := b[i]>>16, b[i]&0xFFFF

			v := tryteRadix*lower + carry
			carry, lower = v>>16, v&0xFFFF
			v = tryteRadix*upper + carry
			carry, upper = v>>16, v&0xFFFF

			b[i] = (upper << 16) | lower
		}
		if carry > 0 && nzIndex < IntLength-1 {
			nzIndex++
			b[nzIndex] = carry
		}

		// then, add the non-balanced tryte value
		chgIndex := bigint.AddSmall(b, uint32(vs[i]+halfTryte))
		// adapt the non-zero index, if we had an overflow
		if chgIndex > nzIndex {
			nzIndex = chgIndex
		}
	}

	// subtract the middle of the domain to get balanced ternary
	bigint.MustSub(b, halfThree)

	bytes := make([]byte, HashBytesSize)
	bytesPutBigint(bytes, b)
	return bytes
}
