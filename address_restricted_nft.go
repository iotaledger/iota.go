package iotago

import (
	"bytes"
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type RestrictedNFTAddress struct {
	NFTID        [NFTAddressBytesLength]byte `serix:"0,mapKey=nftId"`
	Capabilities AddressCapabilitiesBitMask  `serix:"1,mapKey=capabilities,lengthPrefixType=uint8,maxLen=1"`
}

func (addr *RestrictedNFTAddress) Clone() Address {
	cpy := &RestrictedNFTAddress{}
	copy(cpy.NFTID[:], addr.NFTID[:])
	copy(cpy.Capabilities[:], addr.Capabilities[:])

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
		bytes.Equal(addr.Capabilities, otherAddr.Capabilities)
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
		addr.Capabilities.Size()
}

func (addr *RestrictedNFTAddress) CannotReceiveNativeTokens() bool {
	return addr.Capabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedNFTAddress) CannotReceiveMana() bool {
	return addr.Capabilities.CannotReceiveMana()
}

func (addr *RestrictedNFTAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedNFTAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedNFTAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedNFTAddress) CannotReceiveAccountOutputs() bool {
	return addr.Capabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedNFTAddress) CannotReceiveNFTOutputs() bool {
	return addr.Capabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedNFTAddress) CannotReceiveDelegationOutputs() bool {
	return addr.Capabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedNFTAddress) CapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.Capabilities
}

// RestrictedNFTAddressFromOutputID returns the NFT address computed from a given OutputID.
func RestrictedNFTAddressFromOutputID(outputID OutputID,
	canReceiveNativeTokens bool,
	canReceiveMana bool,
	canReceiveOutputsWithTimelockUnlockCondition bool,
	canReceiveOutputsWithExpirationUnlockCondition bool,
	canReceiveOutputsWithStorageDepositReturnUnlockCondition bool,
	canReceiveAccountOutputs bool,
	canReceiveNFTOutputs bool,
	canReceiveDelegationOutputs bool) *RestrictedNFTAddress {

	nftID := blake2b.Sum256(outputID[:])
	addr := &RestrictedNFTAddress{}
	copy(addr.NFTID[:], nftID[:])

	if canReceiveNativeTokens {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveNativeTokensBitIndex)
	}

	if canReceiveMana {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveManaBitIndex)
	}

	if canReceiveOutputsWithTimelockUnlockCondition {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveOutputsWithTimelockUnlockConditionBitIndex)
	}

	if canReceiveOutputsWithExpirationUnlockCondition {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveOutputsWithExpirationUnlockConditionBitIndex)
	}

	if canReceiveOutputsWithStorageDepositReturnUnlockCondition {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
	}

	if canReceiveAccountOutputs {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveAccountOutputsBitIndex)
	}

	if canReceiveNFTOutputs {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveNFTOutputsBitIndex)
	}

	if canReceiveDelegationOutputs {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveDelegationOutputsBitIndex)
	}

	return addr
}
