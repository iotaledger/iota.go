// +build amd64 arm64 mips64

// Package kerl implements the Kerl hashing function.
package kerl

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
)

func chunksToTryteValues(cs []int32) []int8 {
	vs := make([]int8, HashTrytesSize+3)
	for i, c := range cs {
		isNegative := c < 0
		if isNegative {
			c = -c
		}
		for j := 0; j < 5; j++ {
			rem := int8(c % tryteRadix)
			c = c / tryteRadix
			if rem > halfTryte {
				c += 1
				rem -= tryteRadix
			}
			if isNegative {
				rem = -rem
			}
			vs[i*6+j] = rem
		}
		if isNegative {
			c = -c
		}
		vs[i*6+5] = int8(c)
	}
	return vs[0:81]
}

func bytesToTryteValues(bytes []byte) []int8 {
	b := make([]uint32, IntLength)
	bigintPutBytes(b, bytes)

	// the two's complement representation is only correct, if the number fits
	// into 48 bytes, i.e. has the 243th trit set to 0
	bigintZeroLastTrit(b)

	// convert to the unsigned bigint representing non-balanced ternary
	bigint.MustAdd(b, halfThree)

	cs := make([]int32, hashChunkSize)

	// initially, all words of the bigint are non-zero
	nzIndex := IntLength - 1
	for i := 0; i < hashChunkSize-1; i++ {
		// divide the bigint by the radix
		var rem uint32
		for i := nzIndex; i >= 0; i-- {
			v := (uint64(rem) << 32) | uint64(b[i])
			b[i], rem = uint32(v/chunkRadix), uint32(v%chunkRadix)
		}
		cs[i] = int32(rem) - halfChunk

		// decrement index, if the highest considered word of the bigint turned zero
		if nzIndex > 0 && b[nzIndex] == 0 {
			nzIndex--
		}
	}

	// special case for the last tryte, where no further division is necessary
	cs[hashChunkSize-1] = int32(b[0]) - halfChunk

	vs := chunksToTryteValues(cs)
	vs[HashTrytesSize-1] = tryteZeroLastTrit(vs[HashTrytesSize-1])
	return vs
}
