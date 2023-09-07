package iotago

import (
	"bytes"
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type RestrictedNFTAddress struct {
	NFTID               [NFTAddressBytesLength]byte `serix:"0,mapKey=nftId"`
	AllowedCapabilities AddressCapabilitiesBitMask  `serix:"1,mapKey=allowedCapabilities,lengthPrefixType=uint8,maxLen=1"`
}

func (addr *RestrictedNFTAddress) Clone() Address {
	cpy := &RestrictedNFTAddress{}
	copy(cpy.NFTID[:], addr.NFTID[:])
	copy(cpy.AllowedCapabilities[:], addr.AllowedCapabilities[:])

	return cpy
}

func (addr *RestrictedNFTAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *RestrictedNFTAddress) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *RestrictedNFTAddress) Equal(other Address) bool {
	otherAddr, is := other.(*RestrictedNFTAddress)
	if !is {
		return false
	}

	return addr.NFTID == otherAddr.NFTID &&
		bytes.Equal(addr.AllowedCapabilities, otherAddr.AllowedCapabilities)
}

func (addr *RestrictedNFTAddress) Type() AddressType {
	return AddressRestrictedNFT
}

func (addr *RestrictedNFTAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, addr)
}

func (addr *RestrictedNFTAddress) String() string {
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *RestrictedNFTAddress) Size() int {
	return NFTAddressSerializedBytesSize +
		addr.AllowedCapabilities.Size()
}

func (addr *RestrictedNFTAddress) CannotReceiveNativeTokens() bool {
	return addr.AllowedCapabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedNFTAddress) CannotReceiveMana() bool {
	return addr.AllowedCapabilities.CannotReceiveMana()
}

func (addr *RestrictedNFTAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedNFTAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedNFTAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedNFTAddress) CannotReceiveAccountOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedNFTAddress) CannotReceiveNFTOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedNFTAddress) CannotReceiveDelegationOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedNFTAddress) AllowedCapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.AllowedCapabilities
}

func RestrictedNFTAddressFromOutputID(outputID OutputID) *RestrictedNFTAddress {
	nftID := blake2b.Sum256(outputID[:])
	addr := &RestrictedNFTAddress{}
	copy(addr.NFTID[:], nftID[:])

	return addr
}

// RestrictedNFTAddressFromOutputIDWithCapabilities returns the NFT address computed from a given OutputID.
func RestrictedNFTAddressFromOutputIDWithCapabilities(outputID OutputID,
	canReceiveNativeTokens bool,
	canReceiveMana bool,
	canReceiveOutputsWithTimelockUnlockCondition bool,
	canReceiveOutputsWithExpirationUnlockCondition bool,
	canReceiveOutputsWithStorageDepositReturnUnlockCondition bool,
	canReceiveAccountOutputs bool,
	canReceiveNFTOutputs bool,
	canReceiveDelegationOutputs bool) *RestrictedNFTAddress {
	addr := RestrictedNFTAddressFromOutputID(outputID)
	addr.AllowedCapabilities = AddressCapabilitiesBitMaskWithCapabilities(
		canReceiveNativeTokens,
		canReceiveMana,
		canReceiveOutputsWithTimelockUnlockCondition,
		canReceiveOutputsWithExpirationUnlockCondition,
		canReceiveOutputsWithStorageDepositReturnUnlockCondition,
		canReceiveAccountOutputs,
		canReceiveNFTOutputs,
		canReceiveDelegationOutputs,
	)

	return addr
}
