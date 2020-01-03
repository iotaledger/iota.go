// +build !arm,!386,!mips,!mipsle

// Package kerl implements the Kerl hashing function.
package kerl

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
)

func chunksToTryteValues(cs []uint32) []int8 {
	vs := make([]int8, HashTrytesSize)
	for i, c := range cs {
		for j := 0; j < trytesPerChunk-1; j++ {
			rem := int8(c % tryteRadix)
			vs[i*trytesPerChunk+j] = rem - halfTryte
			c = c / tryteRadix

			if i*trytesPerChunk+j >= HashTrytesSize-1 {
				return vs
			}
		}
		vs[i*trytesPerChunk+trytesPerChunk-1] = int8(c) - halfTryte
	}
	return vs
}

func bytesToTryteValues(bytes []byte) []int8 {
	b := make([]uint32, IntLength)
	bigintPutBytes(b, bytes)

	// the two's complement representation is only correct, if the number fits
	// into 48 bytes, i.e. has the 243th trit set to 0
	bigintZeroLastTrit(b)

	// convert to the unsigned bigint representing non-balanced ternary
	bigint.MustAdd(b, halfThree)

	cs := make([]uint32, hashChunkSize)
	// initially, all words of the bigint are non-zero
	nzIndex := IntLength - 1
	for i := 0; i < hashChunkSize-1; i++ {
		// divide the bigint by the radix
		var rem uint32
		for i := nzIndex; i >= 0; i-- {
			v := (uint64(rem) << 32) | uint64(b[i])
			b[i], rem = uint32(v/chunkRadix), uint32(v%chunkRadix)
		}
		cs[i] = rem

		// decrement index, if the highest considered word of the bigint turned zero
		if nzIndex > 0 && b[nzIndex] == 0 {
			nzIndex--
		}
	}

	// special case for the last chunk, where no further division is necessary
	cs[hashChunkSize-1] = b[0]

	// convert to trytes and set the last trit to zero
	vs := chunksToTryteValues(cs)
	vs[HashTrytesSize-1] = tryteValueZeroLastTrit(vs[HashTrytesSize-1])
	return vs
}
