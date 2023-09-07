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

func (nftAddr *NFTAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(nftAddr.Size()))
}

func (nftAddr *NFTAddress) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), nftAddr)))
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
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), nftAddr)))
}

func (nftAddr *NFTAddress) Size() int {
	return NFTAddressSerializedBytesSize
}

func (nftAddr *NFTAddress) CanReceiveNativeTokens() bool {
	return true
}

func (nftAddr *NFTAddress) CanReceiveMana() bool {
	return true
}

func (nftAddr *NFTAddress) CanReceiveOutputsWithTimelockUnlockCondition() bool {
	return true
}

func (nftAddr *NFTAddress) CanReceiveOutputsWithExpirationUnlockCondition() bool {
	return true
}

func (nftAddr *NFTAddress) CanReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return true
}

func (nftAddr *NFTAddress) CanReceiveAccountOutputs() bool {
	return true
}

func (nftAddr *NFTAddress) CanReceiveNFTOutputs() bool {
	return true
}

func (nftAddr *NFTAddress) CanReceiveDelegationOutputs() bool {
	return true
}

// NFTAddressFromOutputID returns the NFT address computed from a given OutputID.
func NFTAddressFromOutputID(outputID OutputID) NFTAddress {
	return blake2b.Sum256(outputID[:])
}
