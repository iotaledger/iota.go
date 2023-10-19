package iotago

import (
	"bytes"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

// MultiAddressReference is a reference to a MultiAddress.
// HINT: This is not an actual AddressType that is used in the protocol, so it should not be registered in serix.
// It should only be used internally or in APIs.
type MultiAddressReference struct {
	MultiAddressID []byte
}

func (addr *MultiAddressReference) Clone() Address {
	return &MultiAddressReference{
		MultiAddressID: lo.CopySlice(addr.MultiAddressID),
	}
}

func (addr *MultiAddressReference) StorageScore(_ *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	panic("not used")
}

func (addr *MultiAddressReference) ID() []byte {
	return addr.MultiAddressID
}

func (addr *MultiAddressReference) Key() string {
	return string(addr.ID())
}

func (addr *MultiAddressReference) Equal(other Address) bool {
	otherAddr, is := other.(*MultiAddressReference)
	if !is {
		return false
	}

	return bytes.Equal(addr.MultiAddressID, otherAddr.MultiAddressID)
}

func (addr *MultiAddressReference) Type() AddressType {
	return AddressMulti
}

func (addr *MultiAddressReference) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *MultiAddressReference) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *MultiAddressReference) Size() int {
	panic("not used")
}

func MultiAddressReferenceFromBytes(bytes []byte) (*MultiAddressReference, int, error) {
	if len(bytes) < AddressMultiIDLength {
		return nil, 0, ierrors.New("invalid multi address ID length")
	}

	if bytes[0] != byte(AddressMulti) {
		return nil, 0, ErrInvalidAddressType
	}

	return &MultiAddressReference{MultiAddressID: bytes[:AddressMultiIDLength]}, AddressMultiIDLength, nil
}

func NewMultiAddressReferenceFromMultiAddress(address *MultiAddress) *MultiAddressReference {
	return &MultiAddressReference{
		MultiAddressID: address.ID(),
	}
}
