package iotago

import (
	"bytes"
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type RestrictedAccountAddress struct {
	AccountID    [AccountAddressBytesLength]byte `serix:"0,mapKey=accountId"`
	Capabilities AddressCapabilitiesBitMask      `serix:"1,mapKey=capabilities,lengthPrefixType=uint8,maxLen=1"`
}

func (addr *RestrictedAccountAddress) Clone() Address {
	cpy := &RestrictedAccountAddress{}
	copy(cpy.AccountID[:], addr.AccountID[:])
	copy(cpy.Capabilities[:], addr.Capabilities[:])

	return cpy
}

func (addr *RestrictedAccountAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *RestrictedAccountAddress) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *RestrictedAccountAddress) Equal(other Address) bool {
	otherAddr, is := other.(*RestrictedAccountAddress)
	if !is {
		return false
	}

	return addr.AccountID == otherAddr.AccountID &&
		bytes.Equal(addr.Capabilities, otherAddr.Capabilities)
}

func (addr *RestrictedAccountAddress) Type() AddressType {
	return AddressRestrictedAccount
}

func (addr *RestrictedAccountAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, addr)
}

func (addr *RestrictedAccountAddress) String() string {
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *RestrictedAccountAddress) Size() int {
	return AccountAddressSerializedBytesSize +
		addr.Capabilities.Size()
}

func (addr *RestrictedAccountAddress) CannotReceiveNativeTokens() bool {
	return addr.Capabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedAccountAddress) CannotReceiveMana() bool {
	return addr.Capabilities.CannotReceiveMana()
}

func (addr *RestrictedAccountAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedAccountAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedAccountAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedAccountAddress) CannotReceiveAccountOutputs() bool {
	return addr.Capabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedAccountAddress) CannotReceiveNFTOutputs() bool {
	return addr.Capabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedAccountAddress) CannotReceiveDelegationOutputs() bool {
	return addr.Capabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedAccountAddress) CapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.Capabilities
}

// RestrictedAccountAddressFromOutputID returns the account address computed from a given OutputID.
func RestrictedAccountAddressFromOutputID(outputID OutputID,
	canReceiveNativeTokens bool,
	canReceiveMana bool,
	canReceiveOutputsWithTimelockUnlockCondition bool,
	canReceiveOutputsWithExpirationUnlockCondition bool,
	canReceiveOutputsWithStorageDepositReturnUnlockCondition bool,
	canReceiveAccountOutputs bool,
	canReceiveNFTOutputs bool,
	canReceiveDelegationOutputs bool) *RestrictedAccountAddress {

	accountID := blake2b.Sum256(outputID[:])
	addr := &RestrictedAccountAddress{}
	copy(addr.AccountID[:], accountID[:])

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
