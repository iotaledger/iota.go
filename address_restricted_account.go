//nolint:dupl // restricted addresses have similar code
package iotago

import (
	"bytes"
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type RestrictedAccountAddress struct {
	AccountID           [AccountAddressBytesLength]byte `serix:"0,mapKey=accountId"`
	AllowedCapabilities AddressCapabilitiesBitMask      `serix:"1,mapKey=allowedCapabilities,lengthPrefixType=uint8,maxLen=1"`
}

func (addr *RestrictedAccountAddress) Clone() Address {
	cpy := &RestrictedAccountAddress{}
	copy(cpy.AccountID[:], addr.AccountID[:])
	copy(cpy.AllowedCapabilities[:], addr.AllowedCapabilities[:])

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
		bytes.Equal(addr.AllowedCapabilities, otherAddr.AllowedCapabilities)
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
		addr.AllowedCapabilities.Size()
}

func (addr *RestrictedAccountAddress) CannotReceiveNativeTokens() bool {
	return addr.AllowedCapabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedAccountAddress) CannotReceiveMana() bool {
	return addr.AllowedCapabilities.CannotReceiveMana()
}

func (addr *RestrictedAccountAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedAccountAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedAccountAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedAccountAddress) CannotReceiveAccountOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedAccountAddress) CannotReceiveNFTOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedAccountAddress) CannotReceiveDelegationOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedAccountAddress) AllowedCapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.AllowedCapabilities
}

// RestrictedAccountAddressFromOutputID returns the account address computed from a given OutputID.
func RestrictedAccountAddressFromOutputID(outputID OutputID) *RestrictedAccountAddress {
	accountID := blake2b.Sum256(outputID[:])
	addr := &RestrictedAccountAddress{}
	copy(addr.AccountID[:], accountID[:])

	return addr
}

// RestrictedAccountAddressFromOutputIDWithCapabilities returns the account address computed from a given OutputID.
func RestrictedAccountAddressFromOutputIDWithCapabilities(outputID OutputID,
	canReceiveNativeTokens bool,
	canReceiveMana bool,
	canReceiveOutputsWithTimelockUnlockCondition bool,
	canReceiveOutputsWithExpirationUnlockCondition bool,
	canReceiveOutputsWithStorageDepositReturnUnlockCondition bool,
	canReceiveAccountOutputs bool,
	canReceiveNFTOutputs bool,
	canReceiveDelegationOutputs bool) *RestrictedAccountAddress {
	addr := RestrictedAccountAddressFromOutputID(outputID)
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
