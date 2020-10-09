package bech32

import (
	"strings"
)

type encoding struct {
	enc    [32]byte
	decMap [256]uint8
}

// newEncoding returns a new encoding defined by the given alphabet,
// which must be a 32-byte string.
func newEncoding(charset string) *encoding {
	if len(charset) != 32 {
		panic("encoding alphabet is not 32-bytes long")
	}

	e := new(encoding)
	copy(e.enc[:], charset)

	for i := 0; i < len(e.decMap); i++ {
		e.decMap[i] = 0xFF
	}
	for i := 0; i < len(charset); i++ {
		e.decMap[charset[i]] = uint8(i)
	}
	return e
}

// encode converts the base32 digits of src into a string.
func (e *encoding) encode(src []uint8) string {
	var dst strings.Builder
	dst.Grow(len(src))
	for i := range src {
		dst.WriteByte(e.enc[src[i]])
	}
	return dst.String()
}

// decode converts the string into base32 digits.
func (e *encoding) decode(src string) ([]uint8, error) {
	dst := make([]uint8, len(src))
	for i := range src {
		d := e.decMap[src[i]]
		if d == 0xFF {
			return dst[:i], ErrInvalidCharacter
		}
		dst[i] = e.decMap[src[i]]
	}
	return dst, nil
}
