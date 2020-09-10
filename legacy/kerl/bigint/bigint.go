// Package bigint contains a very lightweight and high-performance implementation of unsigned multi-precision integers.
package bigint

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/bits"
)

type Word = uint32

const (
	wordSize     = 32
	wordByteSize = wordSize / 8
)

// A Bigint is an unsigned integer x of the form
//   x = x[n-1]*B^(n-1) + x[n-2]*B^(n-2) + ... + x[1]*B + x[0]
// with 0 <= x[i] < B and 0 <= i < n is stored in a slice of length n, with the digits x[i] as the slice elements.
// The length of the slice never changes and operations are only well-defined for operands of the same length.
type Bigint []Word

// Errors for bigint package.
var (
	ErrUnsupportedOperandSize = errors.New("unsupported operand size")
	ErrInvalidBufferSize      = errors.New("invalid buffer sizer")
	ErrInvalidLength          = errors.New("byte length not a multiple of the word size")
	ErrMissingPrefix          = errors.New("hex string without 0x prefix")
)

var hexPrefix = []byte("0x")

// U384 creates a 384-bit big unsigned integer.
func U384() Bigint {
	return make([]Word, 384/wordSize)
}

// ParseU384 parses s as a Bigint, returning the result.
// The encoding of s must be the same as in UnmarshalText.
func ParseU384(s string) (Bigint, error) {
	x := U384()
	if err := x.UnmarshalText([]byte(s)); err != nil {
		return nil, err
	}
	return x, nil
}

// MustParseU384 parses s as a Bigint, returning the result.
// The encoding of s must be the same as in UnmarshalText, it panics when the encoding is not valid.
func MustParseU384(s string) Bigint {
	x, err := ParseU384(s)
	if err != nil {
		panic(err)
	}
	return x
}

// SetBytes interprets buf as the bytes of a big-endian unsigned integer and sets x to that value.
func (x Bigint) SetBytes(buf []byte) {
	n := x.BytesLen()
	if n != len(buf) {
		panic(ErrInvalidBufferSize)
	}
	for i := len(x) - 1; i >= 0; i-- {
		x[i] = binary.BigEndian.Uint32(buf[n-i*wordByteSize-wordByteSize:])
	}
}

// Read reads the big-endian byte representation of x into out and returns the number of bytes read.
// It returns an error when len(out) <= x.BytesLen() and the bytes cannot be written in one call.
func (x Bigint) Read(out []byte) (n int, err error) {
	if len(out) < x.BytesLen() {
		return 0, ErrInvalidBufferSize
	}
	n = x.BytesLen()
	for i := len(x) - 1; i >= 0; i-- {
		binary.BigEndian.PutUint32(out[n-i*wordByteSize-wordByteSize:], x[i])
	}
	return
}

// Words provides raw access to x by returning its underlying little-endian Word slice.
func (x Bigint) Words() []Word {
	return x
}

// BytesLen returns the length of x in bytes.
func (x Bigint) BytesLen() int {
	return len(x) * wordByteSize
}

// MSB returns the value of the most significant bit of x.
func (x Bigint) MSB() uint {
	return uint(x[len(x)-1] >> (wordSize - 1))
}

// Add sets x to the sum x+y and returns the carry.
// The carry output is guaranteed to be 0 or 1.
func (x Bigint) Add(y Bigint) Word {
	if len(x) != len(y) {
		panic(ErrUnsupportedOperandSize)
	}

	var carry Word
	for i := range x {
		x[i], carry = bits.Add32(x[i], y[i], carry)
	}
	return carry
}

// Sub sets x to the difference x-y and returns the borrow.
// The borrow output is guaranteed to be 0 or 1.
func (x Bigint) Sub(y Bigint) Word {
	if len(x) != len(y) {
		panic(ErrUnsupportedOperandSize)
	}

	var borrow Word
	for i := range x {
		x[i], borrow = bits.Sub32(x[i], y[i], borrow)
	}
	return borrow
}

// Cmp compares x and y and returns:
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
func (x Bigint) Cmp(y Bigint) int {
	if len(x) != len(y) {
		panic(ErrUnsupportedOperandSize)
	}

	for i := len(x) - 1; i >= 0; i-- {
		switch {
		case x[i] < y[i]:
			return -1
		case x[i] > y[i]:
			return 1
		}
	}
	return 0
}

// String returns the hexadecimal representation of x including leading zeroes and with the "0x" prefix.
func (x Bigint) String() string {
	text, _ := x.MarshalText()
	return string(text)
}

// MarshalText implements the encoding.TextMarshaler interface.
// The encoding is the same as returned by String.
func (x Bigint) MarshalText() ([]byte, error) {
	buf := make([]byte, x.BytesLen())
	n, err := x.Read(buf)
	if err != nil {
		return nil, err
	}
	text := make([]byte, hex.EncodedLen(n))
	hex.Encode(text, buf[:n])
	return append(hexPrefix, text...), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The Bigint is expected in hexadecimal representation starting with the prefix "0x".
// It must include leading zeroes such that the encoded byte length matches x.BytesLen().
func (x *Bigint) UnmarshalText(text []byte) error {
	if !bytes.HasPrefix(text, hexPrefix) {
		return ErrMissingPrefix
	}
	// ignore prefix
	text = text[len(hexPrefix):]
	buf := make([]byte, hex.DecodedLen(len(text)))
	if _, err := hex.Decode(buf, text); err != nil {
		return err
	}
	if len(buf) != x.BytesLen() {
		return ErrInvalidLength
	}
	x.SetBytes(buf)
	return nil
}
