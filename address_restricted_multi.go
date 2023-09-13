package iotago

import (
	"bytes"
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

// RestrictedMultiAddress defines a multi address that consists of addresses with weights and
// a threshold value that needs to be reached to unlock the multi address.
// It has a capability bitmask that enables additional features.
type RestrictedMultiAddress struct {
	Addresses           AddressesWithWeight        `serix:"0,mapKey=addresses"`
	Threshold           uint16                     `serix:"1,mapKey=threshold"`
	AllowedCapabilities AddressCapabilitiesBitMask `serix:"2,mapKey=allowedCapabilities,lengthPrefixType=uint8,maxLen=1"`
}

func (addr *RestrictedMultiAddress) Clone() Address {
	cpy := &RestrictedMultiAddress{
		Addresses: make(AddressesWithWeight, 0, len(addr.Addresses)),
		Threshold: addr.Threshold,
	}

	for i, address := range addr.Addresses {
		cpy.Addresses[i] = &AddressWithWeight{
			Address: address.Address.Clone(),
			Weight:  address.Weight,
		}
	}

	return cpy
}

func (addr *RestrictedMultiAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *RestrictedMultiAddress) ID() []byte {
	addressBytesHash := blake2b.Sum256(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))

	// prefix the hash of the multi address bytes with the AddressType
	return append([]byte{byte(AddressRestrictedMulti)}, addressBytesHash[:]...)
}

func (addr *RestrictedMultiAddress) Key() string {
	return string(addr.ID())
}

func (addr *RestrictedMultiAddress) Equal(other Address) bool {
	otherAddr, is := other.(*RestrictedMultiAddress)
	if !is {
		return false
	}

	if len(addr.Addresses) != len(otherAddr.Addresses) {
		return false
	}
	if addr.Threshold != otherAddr.Threshold {
		return false
	}

	for i, address := range addr.Addresses {
		if !bytes.Equal(address.Address.ID(), otherAddr.Addresses[i].Address.ID()) {
			return false
		}
		if address.Weight != otherAddr.Addresses[i].Weight {
			return false
		}
	}

	return bytes.Equal(addr.AllowedCapabilities, otherAddr.AllowedCapabilities)
}

func (addr *RestrictedMultiAddress) Type() AddressType {
	return AddressRestrictedMulti
}

func (addr *RestrictedMultiAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *RestrictedMultiAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *RestrictedMultiAddress) CannotReceiveNativeTokens() bool {
	return addr.AllowedCapabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedMultiAddress) CannotReceiveMana() bool {
	return addr.AllowedCapabilities.CannotReceiveMana()
}

func (addr *RestrictedMultiAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedMultiAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedMultiAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedMultiAddress) CannotReceiveAccountOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedMultiAddress) CannotReceiveNFTOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedMultiAddress) CannotReceiveDelegationOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedMultiAddress) AllowedCapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.AllowedCapabilities
}

func (addr *RestrictedMultiAddress) Size() int {
	// Address Type + Addresses Length + Threshold
	sum := serializer.SmallTypeDenotationByteSize + serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize

	for _, address := range addr.Addresses {
		sum += address.Size()
	}

	// AllowedCapabilities
	sum += addr.AllowedCapabilities.Size()

	return sum
}
