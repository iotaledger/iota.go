package iotago

import (
	"bytes"
	"context"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type RestrictedAddress struct {
	Address             Address                    `serix:"0,mapKey=address"`
	AllowedCapabilities AddressCapabilitiesBitMask `serix:"1,mapKey=allowedCapabilities"`
}

func (addr *RestrictedAddress) Clone() Address {
	return &RestrictedAddress{
		Address:             addr.Address.Clone(),
		AllowedCapabilities: addr.AllowedCapabilities.Clone(),
	}
}

func (addr *RestrictedAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return 0
}

func (addr *RestrictedAddress) ID() []byte {
	addressID := addr.Address.ID()
	capabilities := lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr.AllowedCapabilities))

	// prefix the ID of the underlying address with the AddressType, and append the capabilties
	return byteutils.ConcatBytes([]byte{byte(AddressRestricted)}, addressID, capabilities)
}

func (addr *RestrictedAddress) Key() string {
	return string(addr.ID())
}

func (addr *RestrictedAddress) Equal(other Address) bool {
	otherAddr, is := other.(*RestrictedAddress)
	if !is {
		return false
	}

	// check equality of the underlying address and the capabilities
	return addr.Address.Equal(otherAddr.Address) && bytes.Equal(addr.AllowedCapabilities, otherAddr.AllowedCapabilities)
}

func (addr *RestrictedAddress) Type() AddressType {
	return AddressRestricted
}

func (addr *RestrictedAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *RestrictedAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *RestrictedAddress) Size() int {
	// address type + underlying address + capabilities
	return serializer.SmallTypeDenotationByteSize + addr.Address.Size() + addr.AllowedCapabilities.Size()
}

func (addr *RestrictedAddress) CannotReceiveNativeTokens() bool {
	return addr.AllowedCapabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedAddress) CannotReceiveMana() bool {
	return addr.AllowedCapabilities.CannotReceiveMana()
}

func (addr *RestrictedAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedAddress) CannotReceiveAccountOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedAddress) CannotReceiveNFTOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedAddress) CannotReceiveDelegationOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedAddress) AllowedCapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.AllowedCapabilities
}

// RestrictedAddressWithCapabilities returns a restricted address for the given underlying address.
func RestrictedAddressWithCapabilities(address Address, opts ...options.Option[AddressCapabilitiesOptions]) *RestrictedAddress {
	return &RestrictedAddress{
		Address:             address,
		AllowedCapabilities: AddressCapabilitiesBitMaskWithCapabilities(opts...),
	}
}
