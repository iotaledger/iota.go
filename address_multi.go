package iotago

import (
	"bytes"
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	AddressWeightSerializedBytesSize = serializer.OneByte
	AddressMultiIDLength             = serializer.OneByte + blake2b.Size256
)

var (
	ErrMultiAddressWeightInvalid    = ierrors.New("multi address weight invalid")
	ErrMultiAddressThresholdInvalid = ierrors.New("multi address treshold invalid")
)

// AddressWithWeight is an Address with a weight used for threshold calculation in a MultiAddress.
type AddressWithWeight struct {
	Address Address `serix:"0,mapKey=address"`
	Weight  byte    `serix:"1,mapKey=weight"`
}

func (a *AddressWithWeight) Size() int {
	// address + weight
	return a.Address.Size() + AddressWeightSerializedBytesSize
}

// AddressesWithWeight is a set of AddressWithWeight.
type AddressesWithWeight []*AddressWithWeight

// MultiAddress defines a multi address that consists of addresses with weights and
// a threshold value that needs to be reached to unlock the multi address.
type MultiAddress struct {
	Addresses AddressesWithWeight `serix:"0,mapKey=addresses"`
	Threshold uint16              `serix:"1,mapKey=threshold"`
}

func (addr *MultiAddress) Clone() Address {
	cpy := &MultiAddress{
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

func (addr *MultiAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData().Multiply(VBytes(addr.Size()))
}

func (addr *MultiAddress) ID() []byte {
	hash := blake2b.Sum256(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))

	// prefix the hash of the multi address bytes with the AddressType
	return byteutils.ConcatBytes([]byte{byte(AddressMulti)}, hash[:])
}

func (addr *MultiAddress) Key() string {
	return string(addr.ID())
}

func (addr *MultiAddress) Equal(other Address) bool {
	otherAddr, is := other.(*MultiAddress)
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

	return true
}

func (addr *MultiAddress) Type() AddressType {
	return AddressMulti
}

func (addr *MultiAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *MultiAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *MultiAddress) Size() int {
	// Address Type + Addresses Length + Threshold
	sum := serializer.SmallTypeDenotationByteSize + serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize

	for _, address := range addr.Addresses {
		sum += address.Size()
	}

	return sum
}

func NewMultiAddress(addresses AddressesWithWeight, threshold uint16) *MultiAddress {
	return &MultiAddress{
		Addresses: addresses,
		Threshold: threshold,
	}
}
