package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// AliasAddressBytesLength is the length of an Alias address.
	AliasAddressBytesLength = blake2b.Size256
	// AliasAddressSerializedBytesSize is the size of a serialized Alias address with its type denoting byte.
	AliasAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + AliasAddressBytesLength
)

// ParseAliasAddressFromHexString parses the given hex string into an AliasAddress.
func ParseAliasAddressFromHexString(hexAddr string) (*AliasAddress, error) {
	addrBytes, err := DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}
	addr := &AliasAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}

// MustParseAliasAddressFromHexString parses the given hex string into an AliasAddress.
// It panics if the hex address is invalid.
func MustParseAliasAddressFromHexString(hexAddr string) *AliasAddress {
	addr, err := ParseAliasAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// AliasAddress defines an Alias address.
// An AliasAddress is the Blake2b-256 hash of the OutputID which created it.
type AliasAddress [AliasAddressBytesLength]byte

func (aliasAddr *AliasAddress) Decode(b []byte) (int, error) {
	copy(aliasAddr[:], b)
	return AliasAddressSerializedBytesSize - 1, nil
}

func (aliasAddr *AliasAddress) Encode() ([]byte, error) {
	var b [AliasAddressSerializedBytesSize - 1]byte
	copy(b[:], aliasAddr[:])
	return b[:], nil
}

func (aliasAddr *AliasAddress) Clone() Address {
	cpy := &AliasAddress{}
	copy(cpy[:], aliasAddr[:])
	return cpy
}

func (aliasAddr *AliasAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(AliasAddressSerializedBytesSize)
}

func (aliasAddr *AliasAddress) Key() string {
	return string(append([]byte{byte(AddressAlias)}, (*aliasAddr)[:]...))
}

func (aliasAddr *AliasAddress) Chain() ChainID {
	return AliasID(*aliasAddr)
}

func (aliasAddr *AliasAddress) AliasID() AliasID {
	return AliasID(*aliasAddr)
}

func (aliasAddr *AliasAddress) Equal(other Address) bool {
	otherAddr, is := other.(*AliasAddress)
	if !is {
		return false
	}
	return *aliasAddr == *otherAddr
}

func (aliasAddr *AliasAddress) Type() AddressType {
	return AddressAlias
}

func (aliasAddr *AliasAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, aliasAddr)
}

func (aliasAddr *AliasAddress) String() string {
	return EncodeHex(aliasAddr[:])
}

func (aliasAddr *AliasAddress) Size() int {
	return AliasAddressSerializedBytesSize
}

// AliasAddressFromOutputID returns the alias address computed from a given OutputID.
func AliasAddressFromOutputID(outputID OutputID) AliasAddress {
	return blake2b.Sum256(outputID[:])
}
