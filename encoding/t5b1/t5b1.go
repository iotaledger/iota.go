// Package t5b1 implements the t5b1 encoding encoding which uses one byte to represent each 5-trit group.
package t5b1

import (
	"fmt"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

const (
	tritsInByte   = 5
	maxGroupValue = 1 + 3 + 9 + 27 + 81
	minGroupValue = -maxGroupValue
)

// lookup table to unpack a byte into 5 trits.
var tritsLUT = [256][tritsInByte]int8{
	{0, 0, 0, 0, 0}, {1, 0, 0, 0, 0}, {-1, 1, 0, 0, 0}, {0, 1, 0, 0, 0}, {1, 1, 0, 0, 0}, {-1, -1, 1, 0, 0},
	{0, -1, 1, 0, 0}, {1, -1, 1, 0, 0}, {-1, 0, 1, 0, 0}, {0, 0, 1, 0, 0}, {1, 0, 1, 0, 0}, {-1, 1, 1, 0, 0},
	{0, 1, 1, 0, 0}, {1, 1, 1, 0, 0}, {-1, -1, -1, 1, 0}, {0, -1, -1, 1, 0}, {1, -1, -1, 1, 0}, {-1, 0, -1, 1, 0},
	{0, 0, -1, 1, 0}, {1, 0, -1, 1, 0}, {-1, 1, -1, 1, 0}, {0, 1, -1, 1, 0}, {1, 1, -1, 1, 0}, {-1, -1, 0, 1, 0},
	{0, -1, 0, 1, 0}, {1, -1, 0, 1, 0}, {-1, 0, 0, 1, 0}, {0, 0, 0, 1, 0}, {1, 0, 0, 1, 0}, {-1, 1, 0, 1, 0},
	{0, 1, 0, 1, 0}, {1, 1, 0, 1, 0}, {-1, -1, 1, 1, 0}, {0, -1, 1, 1, 0}, {1, -1, 1, 1, 0}, {-1, 0, 1, 1, 0},
	{0, 0, 1, 1, 0}, {1, 0, 1, 1, 0}, {-1, 1, 1, 1, 0}, {0, 1, 1, 1, 0}, {1, 1, 1, 1, 0}, {-1, -1, -1, -1, 1},
	{0, -1, -1, -1, 1}, {1, -1, -1, -1, 1}, {-1, 0, -1, -1, 1}, {0, 0, -1, -1, 1}, {1, 0, -1, -1, 1}, {-1, 1, -1, -1, 1},
	{0, 1, -1, -1, 1}, {1, 1, -1, -1, 1}, {-1, -1, 0, -1, 1}, {0, -1, 0, -1, 1}, {1, -1, 0, -1, 1}, {-1, 0, 0, -1, 1},
	{0, 0, 0, -1, 1}, {1, 0, 0, -1, 1}, {-1, 1, 0, -1, 1}, {0, 1, 0, -1, 1}, {1, 1, 0, -1, 1}, {-1, -1, 1, -1, 1},
	{0, -1, 1, -1, 1}, {1, -1, 1, -1, 1}, {-1, 0, 1, -1, 1}, {0, 0, 1, -1, 1}, {1, 0, 1, -1, 1}, {-1, 1, 1, -1, 1},
	{0, 1, 1, -1, 1}, {1, 1, 1, -1, 1}, {-1, -1, -1, 0, 1}, {0, -1, -1, 0, 1}, {1, -1, -1, 0, 1}, {-1, 0, -1, 0, 1},
	{0, 0, -1, 0, 1}, {1, 0, -1, 0, 1}, {-1, 1, -1, 0, 1}, {0, 1, -1, 0, 1}, {1, 1, -1, 0, 1}, {-1, -1, 0, 0, 1},
	{0, -1, 0, 0, 1}, {1, -1, 0, 0, 1}, {-1, 0, 0, 0, 1}, {0, 0, 0, 0, 1}, {1, 0, 0, 0, 1}, {-1, 1, 0, 0, 1},
	{0, 1, 0, 0, 1}, {1, 1, 0, 0, 1}, {-1, -1, 1, 0, 1}, {0, -1, 1, 0, 1}, {1, -1, 1, 0, 1}, {-1, 0, 1, 0, 1},
	{0, 0, 1, 0, 1}, {1, 0, 1, 0, 1}, {-1, 1, 1, 0, 1}, {0, 1, 1, 0, 1}, {1, 1, 1, 0, 1}, {-1, -1, -1, 1, 1},
	{0, -1, -1, 1, 1}, {1, -1, -1, 1, 1}, {-1, 0, -1, 1, 1}, {0, 0, -1, 1, 1}, {1, 0, -1, 1, 1}, {-1, 1, -1, 1, 1},
	{0, 1, -1, 1, 1}, {1, 1, -1, 1, 1}, {-1, -1, 0, 1, 1}, {0, -1, 0, 1, 1}, {1, -1, 0, 1, 1}, {-1, 0, 0, 1, 1},
	{0, 0, 0, 1, 1}, {1, 0, 0, 1, 1}, {-1, 1, 0, 1, 1}, {0, 1, 0, 1, 1}, {1, 1, 0, 1, 1}, {-1, -1, 1, 1, 1},
	{0, -1, 1, 1, 1}, {1, -1, 1, 1, 1}, {-1, 0, 1, 1, 1}, {0, 0, 1, 1, 1}, {1, 0, 1, 1, 1}, {-1, 1, 1, 1, 1},
	{0, 1, 1, 1, 1}, {1, 1, 1, 1, 1}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {-1, -1, -1, -1, -1}, {0, -1, -1, -1, -1}, {1, -1, -1, -1, -1},
	{-1, 0, -1, -1, -1}, {0, 0, -1, -1, -1}, {1, 0, -1, -1, -1}, {-1, 1, -1, -1, -1}, {0, 1, -1, -1, -1}, {1, 1, -1, -1, -1},
	{-1, -1, 0, -1, -1}, {0, -1, 0, -1, -1}, {1, -1, 0, -1, -1}, {-1, 0, 0, -1, -1}, {0, 0, 0, -1, -1}, {1, 0, 0, -1, -1},
	{-1, 1, 0, -1, -1}, {0, 1, 0, -1, -1}, {1, 1, 0, -1, -1}, {-1, -1, 1, -1, -1}, {0, -1, 1, -1, -1}, {1, -1, 1, -1, -1},
	{-1, 0, 1, -1, -1}, {0, 0, 1, -1, -1}, {1, 0, 1, -1, -1}, {-1, 1, 1, -1, -1}, {0, 1, 1, -1, -1}, {1, 1, 1, -1, -1},
	{-1, -1, -1, 0, -1}, {0, -1, -1, 0, -1}, {1, -1, -1, 0, -1}, {-1, 0, -1, 0, -1}, {0, 0, -1, 0, -1}, {1, 0, -1, 0, -1},
	{-1, 1, -1, 0, -1}, {0, 1, -1, 0, -1}, {1, 1, -1, 0, -1}, {-1, -1, 0, 0, -1}, {0, -1, 0, 0, -1}, {1, -1, 0, 0, -1},
	{-1, 0, 0, 0, -1}, {0, 0, 0, 0, -1}, {1, 0, 0, 0, -1}, {-1, 1, 0, 0, -1}, {0, 1, 0, 0, -1}, {1, 1, 0, 0, -1},
	{-1, -1, 1, 0, -1}, {0, -1, 1, 0, -1}, {1, -1, 1, 0, -1}, {-1, 0, 1, 0, -1}, {0, 0, 1, 0, -1}, {1, 0, 1, 0, -1},
	{-1, 1, 1, 0, -1}, {0, 1, 1, 0, -1}, {1, 1, 1, 0, -1}, {-1, -1, -1, 1, -1}, {0, -1, -1, 1, -1}, {1, -1, -1, 1, -1},
	{-1, 0, -1, 1, -1}, {0, 0, -1, 1, -1}, {1, 0, -1, 1, -1}, {-1, 1, -1, 1, -1}, {0, 1, -1, 1, -1}, {1, 1, -1, 1, -1},
	{-1, -1, 0, 1, -1}, {0, -1, 0, 1, -1}, {1, -1, 0, 1, -1}, {-1, 0, 0, 1, -1}, {0, 0, 0, 1, -1}, {1, 0, 0, 1, -1},
	{-1, 1, 0, 1, -1}, {0, 1, 0, 1, -1}, {1, 1, 0, 1, -1}, {-1, -1, 1, 1, -1}, {0, -1, 1, 1, -1}, {1, -1, 1, 1, -1},
	{-1, 0, 1, 1, -1}, {0, 0, 1, 1, -1}, {1, 0, 1, 1, -1}, {-1, 1, 1, 1, -1}, {0, 1, 1, 1, -1}, {1, 1, 1, 1, -1},
	{-1, -1, -1, -1, 0}, {0, -1, -1, -1, 0}, {1, -1, -1, -1, 0}, {-1, 0, -1, -1, 0}, {0, 0, -1, -1, 0}, {1, 0, -1, -1, 0},
	{-1, 1, -1, -1, 0}, {0, 1, -1, -1, 0}, {1, 1, -1, -1, 0}, {-1, -1, 0, -1, 0}, {0, -1, 0, -1, 0}, {1, -1, 0, -1, 0},
	{-1, 0, 0, -1, 0}, {0, 0, 0, -1, 0}, {1, 0, 0, -1, 0}, {-1, 1, 0, -1, 0}, {0, 1, 0, -1, 0}, {1, 1, 0, -1, 0},
	{-1, -1, 1, -1, 0}, {0, -1, 1, -1, 0}, {1, -1, 1, -1, 0}, {-1, 0, 1, -1, 0}, {0, 0, 1, -1, 0}, {1, 0, 1, -1, 0},
	{-1, 1, 1, -1, 0}, {0, 1, 1, -1, 0}, {1, 1, 1, -1, 0}, {-1, -1, -1, 0, 0}, {0, -1, -1, 0, 0}, {1, -1, -1, 0, 0},
	{-1, 0, -1, 0, 0}, {0, 0, -1, 0, 0}, {1, 0, -1, 0, 0}, {-1, 1, -1, 0, 0}, {0, 1, -1, 0, 0}, {1, 1, -1, 0, 0},
	{-1, -1, 0, 0, 0}, {0, -1, 0, 0, 0}, {1, -1, 0, 0, 0}, {-1, 0, 0, 0, 0},
}

// EncodedLen returns the byte-length of an encoding of n source trits.
func EncodedLen(n int) int { return (n + tritsInByte - 1) / tritsInByte }

// Encode encodes src into EncodedLen(len(src)) bytes of dst. As a convenience, it returns the number of trits written,
// but this value is always EncodedLen(len(src)).
// Encode implements the t5b1 encoding converting a trit string into binary.
// If the length of src is not a multiple of 5, it is padded with zeroes.
// If src does not contain trits, the behavior of Encode is undefined.
func Encode(dst []byte, src trinary.Trits) int {
	i := 0
	for len(src) > 0 {
		// incomplete group
		if len(src) < tritsInByte {
			var v int
			for j := len(src) - 1; j >= 0; j-- {
				v = v*consts.TrinaryRadix + int(src[j])
			}
			dst[i] = byte(v)
			return i + 1
		}
		// common case, unrolled for extra performance
		v := int(src[0]) + int(src[1])*3 + int(src[2])*9 + int(src[3])*27 + int(src[4])*81

		// advance
		dst[i] = byte(v)
		i++
		src = src[tritsInByte:]
	}
	return i
}

// EncodeTrytes returns the encoding of src converted into bytes.
// If the corresponding number of trits of src is not a multiple of 5, it is padded with zeroes.
// If src does not contain valid trytes, the behavior is undefined.
func EncodeTrytes(src trinary.Trytes) []byte {
	dst := make([]byte, EncodedLen(len(src)*consts.TritsPerTryte))
	Encode(dst, trinary.MustTrytesToTrits(src))
	return dst
}

// DecodedLen returns the trit-length of a decoding of n source bytes.
func DecodedLen(n int) int { return n * tritsInByte }

// Decode decodes src into DecodedLen(len(src)) trits of dst and returns the actual number of bytes written.
// Decode expects that src contains a valid t5b1 encoding, it returns an error otherwise.
func Decode(dst trinary.Trits, src []byte) (int, error) {
	j := 0
	for i, b := range src {
		if int8(b) < minGroupValue || int8(b) > maxGroupValue {
			return j, fmt.Errorf("%w: at index %d (byte: %x)", consts.ErrInvalidByte, i, b)
		}
		_ = dst[4] // bounds check hints to compiler
		dst[0] = tritsLUT[b][0]
		dst[1] = tritsLUT[b][1]
		dst[2] = tritsLUT[b][2]
		dst[3] = tritsLUT[b][3]
		dst[4] = tritsLUT[b][4]

		dst = dst[tritsInByte:]
		j += tritsInByte
	}
	return j, nil
}

// DecodeToTrytes returns the trytes represented by the t5b1 encoded bytes. The returned trytes always
// have a length of Ceil(DecodedLen(len(src))/3).
// DecodeToTrytes expects that src contains a valid t5b1 encoding of a tryte-string.
// If the input is malformed or does not contain the correct zero padding, it returns an error.
func DecodeToTrytes(src []byte) (trinary.Trytes, error) {
	trytesLen := (DecodedLen(len(src)) + consts.TritsPerTryte - 1) / consts.TritsPerTryte
	trits := make(trinary.Trits, trytesLen*consts.TritsPerTryte)
	if _, err := Decode(trits, src); err != nil {
		return "", err
	}
	return trinary.MustTritsToTrytes(trits), nil
}
