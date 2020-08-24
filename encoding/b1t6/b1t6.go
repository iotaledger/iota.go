// Package b1t6 implements the b1t6 encoding encoding which uses a group of 6 trits to encode each byte.
// See the IOTA protocol RFC-0015 for details.
package b1t6

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

const (
	tritsPerByte  = 6
	trytesPerByte = tritsPerByte / consts.TritsPerTryte
)

// EncodedLen returns the trit-length of an encoding of n source bytes.
func EncodedLen(n int) int { return n * tritsPerByte }

// Encode encodes src into EncodedLen(len(in)) trits of dst. As a convenience, it returns the number of trits written,
// but this value is always EncodedLen(len(src)).
// Encode implements the b1t6 encoding converting a bit string into ternary.
func Encode(dst trinary.Trits, src []byte) int {
	j := 0
	for i := range src {
		t1, t2 := encodeGroup(src[i])
		trinary.MustPutTryteTrits(dst[j:], t1)
		trinary.MustPutTryteTrits(dst[j+consts.TritsPerTryte:], t2)
		j += 6
	}
	return j
}

// EncodeToTrytes returns the encoding of src converted into trytes.
func EncodeToTrytes(src []byte) trinary.Trytes {
	var dst strings.Builder
	dst.Grow(EncodedLen(len(src)) / consts.TritsPerTryte)

	for i := range src {
		t1, t2 := encodeGroup(src[i])
		dst.WriteByte(trinary.MustTryteValueToTryte(t1))
		dst.WriteByte(trinary.MustTryteValueToTryte(t2))
	}
	return dst.String()
}

var (
	// ErrInvalidLength reports an attempt to decode an input which has a trit-length that is not a multiple of 6.
	ErrInvalidLength = errors.New("length must be a multiple of 6 trits")
	// ErrInvalidTrits reports an attempt to decode an input that contains an invalid trit sequence.
	ErrInvalidTrits = errors.New("invalid trits")
)

// DecodedLen returns the byte-length of a decoding of n source trits.
func DecodedLen(n int) int { return n / tritsPerByte }

// Decode decodes src into DecodedLen(len(in)) bytes of dst and returns the actual number of bytes written.
// Decode expects that src contains a valid b1t6 encoding and that src has a length that is a multiple of 6,
// it returns an error otherwise. If src does not contain trits, the behavior of Decode is undefined.
func Decode(dst []byte, src trinary.Trits) (int, error) {
	i := 0
	for j := 0; j <= len(src)-tritsPerByte; j += tritsPerByte {
		t1 := trinary.MustTritsToTryteValue(src[j:])
		t2 := trinary.MustTritsToTryteValue(src[j+consts.TritsPerTryte:])
		b, ok := decodeGroup(t1, t2)
		if !ok {
			return i, fmt.Errorf("%w: %v", ErrInvalidTrits, src[j:j+6])
		}
		dst[i] = b
		i++
	}
	if len(src)%tritsPerByte != 0 {
		return i, ErrInvalidLength
	}
	return i, nil
}

// DecodeTrytes returns the bytes represented by the t6b1 encoded trytes.
// DecodeTrytes expects that src contains a valid b1t6 encoding and that in has even length,
// it returns an error otherwise. If src does not contain trytes, the behavior of DecodeTrytes is undefined.
func DecodeTrytes(src trinary.Trytes) ([]byte, error) {
	dst := make([]byte, DecodedLen(len(src)*consts.TritsPerTryte))
	i := 0
	for j := 0; j <= len(src)-trytesPerByte; j += trytesPerByte {
		t1 := trinary.MustTryteToTryteValue(src[j])
		t2 := trinary.MustTryteToTryteValue(src[j+1])
		b, ok := decodeGroup(t1, t2)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidTrits, src[j:j+2])
		}
		dst[i] = b
		i++
	}
	if len(src)%trytesPerByte != 0 {
		return nil, ErrInvalidLength
	}
	return dst, nil
}

// encodeGroup converts a byte into two tryte values.
func encodeGroup(b byte) (int8, int8) {
	// this is equivalent to: IntToTrytes(int8(b), 2)
	v := int(int8(b)) + (consts.TryteRadix/2)*consts.TryteRadix + consts.TryteRadix/2 // make un-balanced
	quo, rem := v/consts.TryteRadix, v%consts.TryteRadix
	return int8(rem + consts.MinTryteValue), int8(quo + consts.MinTryteValue)
}

// decodeGroup converts two tryte values into a byte and a success flag.
func decodeGroup(t1, t2 int8) (byte, bool) {
	v := int(t1) + int(t2)*consts.TryteRadix
	if v < math.MinInt8 || v > math.MaxInt8 {
		return 0, false
	}
	return byte(v), true
}
