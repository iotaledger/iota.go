package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// AccountAddressBytesLength is the length of an Account address.
	AccountAddressBytesLength = blake2b.Size256
	// AccountAddressSerializedBytesSize is the size of a serialized Account address with its type denoting byte.
	AccountAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + AccountAddressBytesLength
)

// ParseAccountAddressFromHexString parses the given hex string into an AccountAddress.
func ParseAccountAddressFromHexString(hexAddr string) (*AccountAddress, error) {
	addrBytes, err := hexutil.DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}
	addr := &AccountAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}

// MustParseAccountAddressFromHexString parses the given hex string into an AccountAddress.
// It panics if the hex address is invalid.
func MustParseAccountAddressFromHexString(hexAddr string) *AccountAddress {
	addr, err := ParseAccountAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

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
	return rentStruct.VBFactorData.Multiply(AccountAddressSerializedBytesSize)
}

func (accountAddr *AccountAddress) Key() string {
	return string(append([]byte{byte(AddressAccount)}, (*accountAddr)[:]...))
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
	return hexutil.EncodeHex(accountAddr[:])
}

func (accountAddr *AccountAddress) Size() int {
	return AccountAddressSerializedBytesSize
}

// AccountAddressFromOutputID returns the account address computed from a given OutputID.
func AccountAddressFromOutputID(outputID OutputID) AccountAddress {
	return blake2b.Sum256(outputID[:])
}
