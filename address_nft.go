//nolint:dupl
package iotago

import (
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// NFTAddressBytesLength is the length of an NFT address.
	NFTAddressBytesLength = blake2b.Size256
	// NFTAddressSerializedBytesSize is the size of a serialized NFT address with its type denoting byte.
	NFTAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + NFTAddressBytesLength
)

// NFTAddress defines an NFT address.
// An NFTAddress is the Blake2b-256 hash of the OutputID which created it.
type NFTAddress [NFTAddressBytesLength]byte

func (addr *NFTAddress) Clone() Address {
	cpy := &NFTAddress{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *NFTAddress) StorageScore(_ *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return 0
}

func (addr *NFTAddress) ID() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr))
}

func (addr *NFTAddress) Key() string {
	return string(addr.ID())
}

func (addr *NFTAddress) ChainID() ChainID {
	return NFTID(*addr)
}

func (addr *NFTAddress) NFTID() NFTID {
	return NFTID(*addr)
}

func (addr *NFTAddress) Equal(other Address) bool {
	otherAddr, is := other.(*NFTAddress)
	if !is {
		return false
	}

	return *addr == *otherAddr
}

func (addr *NFTAddress) Type() AddressType {
	return AddressNFT
}

func (addr *NFTAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *NFTAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *NFTAddress) Size() int {
	return NFTAddressSerializedBytesSize
}

// NFTAddressFromOutputID returns the NFT address computed from a given OutputID.
func NFTAddressFromOutputID(outputID OutputID) *NFTAddress {
	address := blake2b.Sum256(outputID[:])

	return (*NFTAddress)(&address)
}
