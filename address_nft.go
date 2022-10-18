package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// NFTAddressBytesLength is the length of an NFT address.
	NFTAddressBytesLength = blake2b.Size256
	// NFTAddressSerializedBytesSize is the size of a serialized NFT address with its type denoting byte.
	NFTAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + NFTAddressBytesLength
)

// ParseNFTAddressFromHexString parses the given hex string into an NFTAddress.
func ParseNFTAddressFromHexString(hexAddr string) (*NFTAddress, error) {
	addrBytes, err := DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}
	addr := &NFTAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}

// MustParseNFTAddressFromHexString parses the given hex string into an NFTAddress.
// It panics if the hex address is invalid.
func MustParseNFTAddressFromHexString(hexAddr string) *NFTAddress {
	addr, err := ParseNFTAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// NFTAddress defines an NFT address.
// An NFTAddress is the Blake2b-256 hash of the OutputID which created it.
type NFTAddress [NFTAddressBytesLength]byte

func (nftAddr *NFTAddress) Decode(b []byte) (int, error) {
	copy(nftAddr[:], b)
	return NFTAddressSerializedBytesSize - 1, nil
}

func (nftAddr *NFTAddress) Encode() ([]byte, error) {
	var b [NFTAddressSerializedBytesSize - 1]byte
	copy(b[:], nftAddr[:])
	return b[:], nil
}

func (nftAddr *NFTAddress) Clone() Address {
	cpy := &NFTAddress{}
	copy(cpy[:], nftAddr[:])
	return cpy
}

func (nftAddr *NFTAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(NFTAddressSerializedBytesSize)
}

func (nftAddr *NFTAddress) Key() string {
	return string(append([]byte{byte(AddressNFT)}, (*nftAddr)[:]...))
}

func (nftAddr *NFTAddress) Chain() ChainID {
	return NFTID(*nftAddr)
}

func (nftAddr *NFTAddress) NFTID() NFTID {
	return NFTID(*nftAddr)
}

func (nftAddr *NFTAddress) Equal(other Address) bool {
	otherAddr, is := other.(*NFTAddress)
	if !is {
		return false
	}
	return *nftAddr == *otherAddr
}

func (nftAddr *NFTAddress) Type() AddressType {
	return AddressNFT
}

func (nftAddr *NFTAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, nftAddr)
}

func (nftAddr *NFTAddress) String() string {
	return EncodeHex(nftAddr[:])
}

func (nftAddr *NFTAddress) Size() int {
	return NFTAddressSerializedBytesSize
}

// NFTAddressFromOutputID returns the NFT address computed from a given OutputID.
func NFTAddressFromOutputID(outputID OutputID) NFTAddress {
	return blake2b.Sum256(outputID[:])
}
