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
	// AccountAddressBytesLength is the length of an Account address.
	AccountAddressBytesLength = blake2b.Size256
	// AccountAddressSerializedBytesSize is the size of a serialized Account address with its type denoting byte.
	AccountAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + AccountAddressBytesLength
)

// AccountAddress defines an Account address.
// An AccountAddress is the Blake2b-256 hash of the OutputID which created it.
type AccountAddress [AccountAddressBytesLength]byte

func (addr *AccountAddress) Clone() Address {
	cpy := &AccountAddress{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *AccountAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *AccountAddress) ID() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr))
}

func (addr *AccountAddress) Key() string {
	return string(addr.ID())
}

func (addr *AccountAddress) ChainID() ChainID {
	return AccountID(*addr)
}

func (addr *AccountAddress) AccountID() AccountID {
	return AccountID(*addr)
}

func (addr *AccountAddress) Equal(other Address) bool {
	otherAddr, is := other.(*AccountAddress)
	if !is {
		return false
	}

	return *addr == *otherAddr
}

func (addr *AccountAddress) Type() AddressType {
	return AddressAccount
}

func (addr *AccountAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *AccountAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *AccountAddress) Size() int {
	return AccountAddressSerializedBytesSize
}

// AccountAddressFromOutputID returns the account address computed from a given OutputID.
func AccountAddressFromOutputID(outputID OutputID) *AccountAddress {
	address := blake2b.Sum256(outputID[:])

	return (*AccountAddress)(&address)
}
