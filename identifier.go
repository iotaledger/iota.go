package iotago

import (
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/blake2b"
)

const (
	// IdentifierLength defines the length of an Identifier.
	IdentifierLength = blake2b.Size256
)

var (
	emptyIdentifier = Identifier{}

	ErrInvalidIdentifierLength = errors.New("Invalid identifier length")
)

// Identifier is a 32 byte hash value that can be used to uniquely identify some blob of data.
type Identifier [IdentifierLength]byte

// IdentifierFromData returns a new Identifier for the given data by hashing it with blake2b.
func IdentifierFromData(data []byte) Identifier {
	return blake2b.Sum256(data)
}

// IdentifierFromHexString converts the hex to an Identifier representation.
func IdentifierFromHexString(hex string) (Identifier, error) {
	bytes, err := DecodeHex(hex)
	if err != nil {
		return Identifier{}, err
	}

	if len(bytes) != IdentifierLength {
		return Identifier{}, ErrInvalidIdentifierLength
	}

	var id Identifier
	copy(id[:], bytes)
	return id, nil
}

// MustIdentifierFromHexString converts the hex to an Identifier representation.
func MustIdentifierFromHexString(hex string) Identifier {
	id, err := IdentifierFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return id
}

func (id Identifier) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(Identifier{})))
	hex.Encode(dst, id[:])
	return dst, nil
}

func (id *Identifier) UnmarshalText(text []byte) error {
	_, err := hex.Decode(id[:], text)
	return err
}

// Empty tells whether the Identifier is empty.
func (id Identifier) Empty() bool {
	return id == emptyIdentifier
}

// ToHex converts the Identifier to its hex representation.
func (id Identifier) ToHex() string {
	return EncodeHex(id[:])
}

func (id Identifier) String() string {
	return id.ToHex()
}
