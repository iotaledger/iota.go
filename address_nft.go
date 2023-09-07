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

func (addr *NFTAddress) Decode(b []byte) (int, error) {
	copy(addr[:], b)

	return NFTAddressSerializedBytesSize - 1, nil
}

func (addr *NFTAddress) Encode() ([]byte, error) {
	var b [NFTAddressSerializedBytesSize - 1]byte
	copy(b[:], addr[:])

	return b[:], nil
}

func (addr *NFTAddress) Clone() Address {
	cpy := &NFTAddress{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *NFTAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *NFTAddress) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *NFTAddress) Chain() ChainID {
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
	return bech32String(hrp, addr)
}

func (addr *NFTAddress) String() string {
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *NFTAddress) Size() int {
	return NFTAddressSerializedBytesSize
}

func (addr *NFTAddress) CanReceiveNativeTokens() bool {
	return true
}

func (addr *NFTAddress) CanReceiveMana() bool {
	return true
}

func (addr *NFTAddress) CanReceiveOutputsWithTimelockUnlockCondition() bool {
	return true
}

func (addr *NFTAddress) CanReceiveOutputsWithExpirationUnlockCondition() bool {
	return true
}

func (addr *NFTAddress) CanReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return true
}

func (addr *NFTAddress) CanReceiveAccountOutputs() bool {
	return true
}

func (addr *NFTAddress) CanReceiveNFTOutputs() bool {
	return true
}

func (addr *NFTAddress) CanReceiveDelegationOutputs() bool {
	return true
}

// NFTAddressFromOutputID returns the NFT address computed from a given OutputID.
func NFTAddressFromOutputID(outputID OutputID) NFTAddress {
	return blake2b.Sum256(outputID[:])
}
