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
	// AnchorAddressBytesLength is the length of an Anchor address.
	AnchorAddressBytesLength = blake2b.Size256
	// AnchorAddressSerializedBytesSize is the size of a serialized Anchor address with its type denoting byte.
	AnchorAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + AnchorAddressBytesLength
)

// AnchorAddress defines an Anchor address.
// An AnchorAddress is the Blake2b-256 hash of the OutputID which created it.
type AnchorAddress [AnchorAddressBytesLength]byte

func (addr *AnchorAddress) Clone() Address {
	cpy := &AnchorAddress{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *AnchorAddress) StorageScore(_ *RentStructure, _ StorageScoreFunc) StorageScore {
	return 0
}

func (addr *AnchorAddress) ID() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr))
}

func (addr *AnchorAddress) Key() string {
	return string(addr.ID())
}

func (addr *AnchorAddress) ChainID() ChainID {
	return AnchorID(*addr)
}

func (addr *AnchorAddress) AnchorID() AnchorID {
	return AnchorID(*addr)
}

func (addr *AnchorAddress) Equal(other Address) bool {
	otherAddr, is := other.(*AnchorAddress)
	if !is {
		return false
	}

	return *addr == *otherAddr
}

func (addr *AnchorAddress) Type() AddressType {
	return AddressAnchor
}

func (addr *AnchorAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *AnchorAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *AnchorAddress) Size() int {
	return AnchorAddressSerializedBytesSize
}

// AnchorAddressFromOutputID returns the Anchor address computed from a given OutputID.
func AnchorAddressFromOutputID(outputID OutputID) *AnchorAddress {
	address := blake2b.Sum256(outputID[:])

	return (*AnchorAddress)(&address)
}
