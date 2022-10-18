package iotago

import (
	"crypto/ed25519"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// Ed25519AddressBytesLength is the length of an Ed25519 address.
	Ed25519AddressBytesLength = blake2b.Size256
	// Ed25519AddressSerializedBytesSize is the size of a serialized Ed25519 address with its type denoting byte.
	Ed25519AddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + Ed25519AddressBytesLength
)

// ParseEd25519AddressFromHexString parses the given hex string into an Ed25519Address.
func ParseEd25519AddressFromHexString(hexAddr string) (*Ed25519Address, error) {
	addrBytes, err := DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}
	addr := &Ed25519Address{}
	copy(addr[:], addrBytes)
	return addr, nil
}

// MustParseEd25519AddressFromHexString parses the given hex string into an Ed25519Address.
// It panics if the hex address is invalid.
func MustParseEd25519AddressFromHexString(hexAddr string) *Ed25519Address {
	addr, err := ParseEd25519AddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// Ed25519Address defines an Ed25519 address.
// An Ed25519Address is the Blake2b-256 hash of an Ed25519 public key.
type Ed25519Address [Ed25519AddressBytesLength]byte

func (edAddr *Ed25519Address) Decode(b []byte) (int, error) {
	copy(edAddr[:], b)
	return Ed25519AddressSerializedBytesSize - 1, nil
}

func (edAddr *Ed25519Address) Encode() ([]byte, error) {
	var b [Ed25519AddressSerializedBytesSize - 1]byte
	copy(b[:], edAddr[:])
	return b[:], nil
}

func (edAddr *Ed25519Address) Clone() Address {
	cpy := &Ed25519Address{}
	copy(cpy[:], edAddr[:])
	return cpy
}

func (edAddr *Ed25519Address) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(Ed25519AddressSerializedBytesSize)
}

func (edAddr *Ed25519Address) Key() string {
	return string(append([]byte{byte(AddressEd25519)}, (*edAddr)[:]...))
}

func (edAddr *Ed25519Address) Unlock(msg []byte, sig Signature) error {
	edSig, isEdSig := sig.(*Ed25519Signature)
	if !isEdSig {
		return fmt.Errorf("%w: can not unlock Ed25519 address with signature of type %s", ErrSignatureAndAddrIncompatible, sig.Type())
	}
	return edSig.Valid(msg, edAddr)
}

func (edAddr *Ed25519Address) Equal(other Address) bool {
	otherAddr, is := other.(*Ed25519Address)
	if !is {
		return false
	}
	return *edAddr == *otherAddr
}

func (edAddr *Ed25519Address) Type() AddressType {
	return AddressEd25519
}

func (edAddr *Ed25519Address) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, edAddr)
}

func (edAddr *Ed25519Address) String() string {
	return EncodeHex(edAddr[:])
}

func (edAddr *Ed25519Address) Size() int {
	return Ed25519AddressSerializedBytesSize
}

// Ed25519AddressFromPubKey returns the address belonging to the given Ed25519 public key.
func Ed25519AddressFromPubKey(pubKey ed25519.PublicKey) Ed25519Address {
	return blake2b.Sum256(pubKey[:])
}
