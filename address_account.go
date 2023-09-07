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

func (addr *AccountAddress) Decode(b []byte) (int, error) {
	copy(addr[:], b)

	return AccountAddressSerializedBytesSize - 1, nil
}

func (addr *AccountAddress) Encode() ([]byte, error) {
	var b [AccountAddressSerializedBytesSize - 1]byte
	copy(b[:], addr[:])

	return b[:], nil
}

func (addr *AccountAddress) Clone() Address {
	cpy := &AccountAddress{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *AccountAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *AccountAddress) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *AccountAddress) Chain() ChainID {
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
	return bech32String(hrp, addr)
}

func (addr *AccountAddress) String() string {
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *AccountAddress) Size() int {
	return AccountAddressSerializedBytesSize
}

func (addr *AccountAddress) CanReceiveNativeTokens() bool {
	return true
}

func (addr *AccountAddress) CanReceiveMana() bool {
	return true
}

func (addr *AccountAddress) CanReceiveOutputsWithTimelockUnlockCondition() bool {
	return true
}

func (addr *AccountAddress) CanReceiveOutputsWithExpirationUnlockCondition() bool {
	return true
}

func (addr *AccountAddress) CanReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return true
}

func (addr *AccountAddress) CanReceiveAccountOutputs() bool {
	return true
}

func (addr *AccountAddress) CanReceiveNFTOutputs() bool {
	return true
}

func (addr *AccountAddress) CanReceiveDelegationOutputs() bool {
	return true
}

// AccountAddressFromOutputID returns the account address computed from a given OutputID.
func AccountAddressFromOutputID(outputID OutputID) AccountAddress {
	return blake2b.Sum256(outputID[:])
}
