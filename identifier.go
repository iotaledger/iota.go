package iotago

import (
	"golang.org/x/crypto/blake2b"
)

const (
	// IdentifierLength defines the length of an Identifier.
	IdentifierLength = 32
)

// Identifier is a 32 byte hash value that can be used to uniquely identify some blob of data.
type Identifier [IdentifierLength]byte

// IdentifierFromData returns a new Identifier for the given data.
func IdentifierFromData(data []byte) Identifier {
	return blake2b.Sum256(data)
}

// IdentifierFromHex creates an Identifier from the given hex encoded data.
func IdentifierFromHex(hexStr string) (Identifier, error) {
	var identifier Identifier
	identifierData, err := DecodeHex(hexStr)
	if err != nil {
		return identifier, err
	}
	copy(identifier[:], identifierData)
	return identifier, nil
}

// ToHex converts the OutputID to its hex representation.
func (identifier Identifier) ToHex() string {
	return EncodeHex(identifier[:])
}
