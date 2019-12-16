// +build amd64 arm64 mips64

// Package kerl implements the Kerl hashing function.
package kerl

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
)

const (
	// largest number of trytes that can be represented as a uint32
	trytesPerChunk = 6
	// radix used in the chunk conversion, i.e. 27^trytesPerChunk
	chunkRadix = 387420489
	// number of chunks to represent the hash, i.e. ceil(HashTrytesSize / trytesPerChunk)
	hashChunkSize = (HashTrytesSize + trytesPerChunk - 1) / trytesPerChunk
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

func tryteValuesToChunk(vs []int8) []uint32 {
	cs := make([]uint32, hashChunkSize)
	for i := 0; i < hashChunkSize; i++ {
		for j := trytesPerChunk - 1; j >= 0; j-- {
			if i*trytesPerChunk+j < HashTrytesSize {
				v := uint32(vs[i*trytesPerChunk+j] + halfTryte)
				cs[i] = cs[i]*tryteRadix + v
			}
		}
	}
	return cs
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

func tryteValuesToBytes(vs []int8) []byte {
	// set the last trit to zero and shift to accommodate for the fact that only 2 trits are used
	vs[HashTrytesSize-1] = tryteValueZeroLastTrit(vs[HashTrytesSize-1]) + 4 - halfTryte
	cs := tryteValuesToChunk(vs)

	b := make([]uint32, IntLength)
	// no multiplication needed for the first chunk
	b[0] = cs[hashChunkSize-1]

	// initially, only the word with index 0 of the bigint is non-zero
	var nzIndex = 0
	for i := hashChunkSize - 2; i >= 0; i-- {
		// multiply the entire bigint by the radix
		var carry uint32
		for i := 0; i <= nzIndex; i++ {
			v := chunkRadix*uint64(b[i]) + uint64(carry)
			carry, b[i] = uint32(v>>32), uint32(v)
		}
		if carry > 0 && nzIndex < IntLength-1 {
			nzIndex++
			b[nzIndex] = carry
		}

		// add the current chunk to the bigint and adapt the non-zero index, if we had an overflow
		chgIndex := bigint.AddSmall(b, cs[i])
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
