// Package kerl implements the Kerl hashing function.
package kerl

import (
	"encoding/binary"
	"strings"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

const (
	// radix used in the conversion
	tryteRadix = 27
	// the middle of the domain described by one tryte
	halfTryte = tryteRadix / 2

	// largest number of trytes that can be represented as a uint32
	trytesPerChunk = 6
	// radix used in the chunk conversion, i.e. 27^trytesPerChunk
	chunkRadix = 387420489
	// number of chunks to represent the hash, i.e. ceil(HashTrytesSize / trytesPerChunk)
	hashChunkSize = (HashTrytesSize + trytesPerChunk - 1) / trytesPerChunk
)

// hex representation of the middle of the domain described by 242 trits, i.e. \sum_{k=0}^{241} 3^k
var halfThree = []uint32{
	0xa5ce8964, 0x9f007669, 0x1484504f, 0x3ade00d9, 0x0c24486e, 0x50979d57,
	0x79a4c702, 0x48bbae36, 0xa9f6808b, 0xaa06a805, 0xa87fabdf, 0x5e69ebef,
}

// hex representation of the two's complement of halfThree, i.e. ~halfThree + 1
var negHalfThree = []uint32{
	0x5a31769c, 0x60ff8996, 0xeb7bafb0, 0xc521ff26, 0xf3dbb791, 0xaf6862a8,
	0x865b38fd, 0xb74451c9, 0x56097f74, 0x55f957fa, 0x57805420, 0xa1961410,
}

// hex representation of the last trit, i.e. 3^242
var trit243 = []uint32{
	0x4b9d12c9, 0x3e00ecd3, 0x2908a09f, 0x75bc01b2, 0x184890dc, 0xa12f3aae,
	0xf3498e04, 0x91775c6c, 0x53ed0116, 0x540d500b, 0x50ff57bf, 0xbcd3d7df,
}

func tryteValuesToTrits(vs []int8) Trits {
	trits := make([]int8, len(vs)*TritsPerTryte)
	for i, v := range vs {
		idx := v - MinTryteValue
		trits[i*TritsPerTryte+0] = TryteValueToTritsLUT[idx][0]
		trits[i*TritsPerTryte+1] = TryteValueToTritsLUT[idx][1]
		trits[i*TritsPerTryte+2] = TryteValueToTritsLUT[idx][2]
	}
	return trits
}

func tryteValuesToTrytes(vs []int8) Trytes {
	var trytes strings.Builder
	trytes.Grow(len(vs))
	for _, v := range vs {
		idx := v - MinTryteValue
		trytes.WriteByte(TryteValueToTyteLUT[idx])
	}
	return trytes.String()
}

func tritsToTryteValues(trits Trits) []int8 {
	vs := make([]int8, len(trits)/TritsPerTryte)
	for i := 0; i < len(trits)/TritsPerTryte; i++ {
		vs[i] = trits[i*TritsPerTryte] + trits[i*TritsPerTryte+1]*3 + trits[i*TritsPerTryte+2]*9
	}
	return vs
}

func trytesToTryteValues(trytes Trytes) []int8 {
	vs := make([]int8, len(trytes))
	for i, tryte := range trytes {
		idx := tryte - '9'
		vs[i] = TryteToTryteValueLUT[idx]
	}
	return vs
}

// tryteValueZeroLastTrit takes a tryte value of three trits a+3b+9c and returns a+3b (setting the last trit to zero).
func tryteValueZeroLastTrit(v int8) int8 {
	if v > 4 {
		return v - 9
	}
	if v < -4 {
		return v + 9
	}
	return v
}

// bigintZeroLastTrit changes the bigint so that the corresponding ternary number has 242th trit set to 0.
// It returns whether the provided bigint was changed.
func bigintZeroLastTrit(b []uint32) bool {
	if bigint.IsNegative(b) {
		if bigint.MustCmp(b, negHalfThree) < 0 {
			bigint.MustAdd(b, trit243)
			return true
		}
	} else {
		if bigint.MustCmp(b, halfThree) > 0 {
			bigint.MustSub(b, trit243)
			return true
		}
	}
	return false
}

// bigintPutBytes decodes the bytes as a bigint in big-endian.
func bigintPutBytes(b []uint32, bytes []byte) {
	for i := 0; i < IntLength; i++ {
		b[IntLength-i-1] = binary.BigEndian.Uint32(bytes[i*4:])
	}
}

// bytesPutBigint encodes the bigint as 48 bytes in big-endian.
func bytesPutBigint(bytes []byte, b []uint32) {
	for i := 0; i < IntLength; i++ {
		binary.BigEndian.PutUint32(bytes[i*4:], b[IntLength-i-1])
	}
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

// KerlBytesZeroLastTrit changes a chunk of 48 bytes so that the corresponding ternary number has 242th trit set to 0.
func KerlBytesZeroLastTrit(bytes []byte) {
	b := make([]uint32, IntLength)
	bigintPutBytes(b, bytes)
	if bigintZeroLastTrit(b) {
		bytesPutBigint(bytes, b)
	}
}

// KerlTritsToBytes is only defined for hashes, i.e. chunks of trits of length 243. It returns 48 bytes.
func KerlTritsToBytes(trits Trits) ([]byte, error) {
	if !CanBeHash(trits) {
		return nil, errors.Wrapf(ErrInvalidTritsLength, "must be %d in size", HashTrinarySize)
	}

	vs := tritsToTryteValues(trits)
	return tryteValuesToBytes(vs), nil
}

// KerlTrytesToBytes is only defined for hashes, i.e. chunks of trytes of length 81. It returns 48 bytes.
func KerlTrytesToBytes(trytes Trytes) ([]byte, error) {
	if len(trytes) != HashTrytesSize {
		return nil, errors.Wrapf(ErrInvalidTrytesLength, "must be %d in size", HashBytesSize)
	}

	vs := trytesToTryteValues(trytes)
	return tryteValuesToBytes(vs), nil
}

// KerlBytesToTrits is only defined for hashes, i.e. chunks of 48 bytes. It returns 243 trits.
func KerlBytesToTrits(b []byte) (Trits, error) {
	if len(b) != HashBytesSize {
		return nil, errors.Wrapf(ErrInvalidBytesLength, "must be %d in size", HashBytesSize)
	}

	vs := bytesToTryteValues(b)
	return tryteValuesToTrits(vs), nil
}

// KerlBytesToTrytes is only defined for hashes, i.e. chunks of 48 bytes. It returns 81 trytes.
func KerlBytesToTrytes(b []byte) (Trytes, error) {
	if len(b) != HashBytesSize {
		return "", errors.Wrapf(ErrInvalidBytesLength, "must be %d in size", HashBytesSize)
	}

	vs := bytesToTryteValues(b)
	return tryteValuesToTrytes(vs), nil
}
