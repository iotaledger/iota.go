// Package base32 implements the conversion for bytes (base256) to base32.
package base32

import (
	"errors"
	"fmt"
)

// EncodedLen returns the length of the base32 encoding of an input buffer of length n.
func EncodedLen(n int) int {
	return (n*8 + 4) / 5
}

// Encode encodes src into EncodedLen(len(src)) digits of dst.
// As a convenience, it returns the number of digits written to dst,
// but this value is always EncodedLen(len(src)).
// Encode implements base32 encoding.
func Encode(dst []uint8, src []byte) int {
	n := EncodedLen(len(src))
	for len(src) > 0 {
		var carry byte

		// unpack 8x 5-bit source blocks into a 5 byte destination quantum
		switch len(src) {
		default:
			dst[7] = src[4] & 0x1F
			carry = src[4] >> 5
			fallthrough
		case 4:
			dst[6] = carry | (src[3]<<3)&0x1F
			dst[5] = (src[3] >> 2) & 0x1F
			carry = src[3] >> 7
			fallthrough
		case 3:
			dst[4] = carry | (src[2]<<1)&0x1F
			carry = (src[2] >> 4) & 0x1F
			fallthrough
		case 2:
			dst[3] = carry | (src[1]<<4)&0x1F
			dst[2] = (src[1] >> 1) & 0x1F
			carry = (src[1] >> 6) & 0x1F
			fallthrough
		case 1:
			dst[1] = carry | (src[0]<<2)&0x1F
			dst[0] = src[0] >> 3
		}

		if len(src) < 5 {
			break
		}
		src = src[5:]
		dst = dst[8:]
	}
	return n
}

var (
	// ErrInvalidLength reports an attempt to decode an input of invalid length.
	ErrInvalidLength = errors.New("invalid length")
	// ErrNonZeroPadding reports an attempt to decode an input without zero padding.
	ErrNonZeroPadding = errors.New("non-zero padding")
)

// A CorruptInputError is a description of a base32 syntax error.
type CorruptInputError struct {
	err    error // wrapped error
	Offset int   // error occurred after reading Offset bytes
}

func (e CorruptInputError) Error() string {
	return fmt.Sprintf("%s at input byte %d", e.err, e.Offset)
}

func (e CorruptInputError) Unwrap() error {
	return e.err
}

// DecodedLen returns the maximum length in bytes of the decoded data corresponding to n base32-encoded values.
func DecodedLen(n int) int {
	return n * 5 / 8
}

// Decode decodes src into DecodedLen(len(src)) bytes, returning the actual number of bytes written to dst.
// If the input is malformed, Decode returns an error and the number of bytes decoded before the error.
func Decode(dst []byte, src []uint8) (int, error) {
	written := 0
	read := 0
	for len(src) > 0 {
		n := len(src)
		if n == 1 || n == 3 || n == 6 {
			return written, &CorruptInputError{ErrInvalidLength, read}
		}

		// pack 8x 5-bit source blocks into 5 byte destination quantum
		switch n {
		default:
			dst[4] = src[6]<<5 | src[7]
			written++
			fallthrough
		case 7:
			dst[3] = src[4]<<7 | src[5]<<2 | src[6]>>3
			written++
			fallthrough
		case 5:
			dst[2] = src[3]<<4 | src[4]>>1
			written++
			fallthrough
		case 4:
			dst[1] = src[1]<<6 | src[2]<<1 | src[3]>>4
			written++
			fallthrough
		case 2:
			dst[0] = src[0]<<3 | src[1]>>2
			written++
		}

		if n < 8 {
			// check for non-zero padding
			switch {
			case n == 2 && src[1]&(1<<2-1) != 0:
				return written, &CorruptInputError{ErrNonZeroPadding, read + 1}
			case n == 4 && src[3]&(1<<4-1) != 0:
				return written, &CorruptInputError{ErrNonZeroPadding, read + 3}
			case n == 5 && src[4]&(1<<1-1) != 0:
				return written, &CorruptInputError{ErrNonZeroPadding, read + 4}
			case n == 7 && src[6]&(1<<3-1) != 0:
				return written, &CorruptInputError{ErrNonZeroPadding, read + 6}
			}
			break
		}
		dst = dst[5:]
		src = src[8:]
		read += 8
	}
	return written, nil
}
