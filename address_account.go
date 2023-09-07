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

func (accountAddr *AccountAddress) Decode(b []byte) (int, error) {
	copy(accountAddr[:], b)

	return AccountAddressSerializedBytesSize - 1, nil
}

func (accountAddr *AccountAddress) Encode() ([]byte, error) {
	var b [AccountAddressSerializedBytesSize - 1]byte
	copy(b[:], accountAddr[:])

	return b[:], nil
}

func (accountAddr *AccountAddress) Clone() Address {
	cpy := &AccountAddress{}
	copy(cpy[:], accountAddr[:])

	return cpy
}

func (accountAddr *AccountAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(accountAddr.Size()))
}

func (accountAddr *AccountAddress) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), accountAddr)))
}

func (accountAddr *AccountAddress) Chain() ChainID {
	return AccountID(*accountAddr)
}

func (accountAddr *AccountAddress) AccountID() AccountID {
	return AccountID(*accountAddr)
}

func (accountAddr *AccountAddress) Equal(other Address) bool {
	otherAddr, is := other.(*AccountAddress)
	if !is {
		return false
	}

	return *accountAddr == *otherAddr
}

func (accountAddr *AccountAddress) Type() AddressType {
	return AddressAccount
}

func (accountAddr *AccountAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, accountAddr)
}

func (accountAddr *AccountAddress) String() string {
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), accountAddr)))
}

func (accountAddr *AccountAddress) Size() int {
	return AccountAddressSerializedBytesSize
}

func (accountAddr *AccountAddress) CanReceiveNativeTokens() bool {
	return true
}

func (accountAddr *AccountAddress) CanReceiveMana() bool {
	return true
}

func (accountAddr *AccountAddress) CanReceiveOutputsWithTimelockUnlockCondition() bool {
	return true
}

func (accountAddr *AccountAddress) CanReceiveOutputsWithExpirationUnlockCondition() bool {
	return true
}

func (accountAddr *AccountAddress) CanReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return true
}

func (accountAddr *AccountAddress) CanReceiveAccountOutputs() bool {
	return true
}

func (accountAddr *AccountAddress) CanReceiveNFTOutputs() bool {
	return true
}

func (accountAddr *AccountAddress) CanReceiveDelegationOutputs() bool {
	return true
}

// AccountAddressFromOutputID returns the account address computed from a given OutputID.
func AccountAddressFromOutputID(outputID OutputID) AccountAddress {
	return blake2b.Sum256(outputID[:])
}
