package tpkg

import (
	"encoding/hex"
)

// HexBytes is a slice of bytes that marshals/unmarshals as a string in hexadecimal encoding.
// It is a simple utility to parse hex encoded test vectors.
type HexBytes []byte

// MarshalText implements the encoding.TextMarshaler interface.
func (b HexBytes) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (b *HexBytes) UnmarshalText(text []byte) (err error) {
	dec := make([]byte, hex.DecodedLen(len(text)))
	if _, err = hex.Decode(dec, text); err != nil {
		return err
	}
	*b = dec
	return
}

// String returns the hex encoding of b.
func (b HexBytes) String() string {
	return hex.EncodeToString(b)
}
