package iotago

// Code generated by go generate; DO NOT EDIT. Check gen/ directory instead.

import (
	"context"
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// ImplicitAccountCreationAddressBytesLength is the length of an ImplicitAccountCreationAddress.
	ImplicitAccountCreationAddressBytesLength = blake2b.Size256
	// ImplicitAccountCreationAddressSerializedBytesSize is the size of a serialized ImplicitAccountCreationAddress with its type denoting byte.
	ImplicitAccountCreationAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + ImplicitAccountCreationAddressBytesLength
)

// ImplicitAccountCreationAddress defines an ImplicitAccountCreationAddress.
// An ImplicitAccountCreationAddress is an address that is used to create implicit accounts by sending basic outputs to it.
type ImplicitAccountCreationAddress [ImplicitAccountCreationAddressBytesLength]byte

func (addr *ImplicitAccountCreationAddress) Clone() Address {
	cpy := &ImplicitAccountCreationAddress{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *ImplicitAccountCreationAddress) ID() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr))
}

func (addr *ImplicitAccountCreationAddress) Key() string {
	return string(addr.ID())
}

func (addr *ImplicitAccountCreationAddress) Unlock(msg []byte, sig Signature) error {
	edSig, isEdSig := sig.(*Ed25519Signature)
	if !isEdSig {
		return ierrors.Wrapf(ErrSignatureAndAddrIncompatible, "can not unlock ImplicitAccountCreationAddress with signature of type %s", sig.Type())
	}

	return edSig.Valid(msg, (*Ed25519Address)(addr))
}

func (addr *ImplicitAccountCreationAddress) Equal(other Address) bool {
	otherAddr, is := other.(*ImplicitAccountCreationAddress)
	if !is {
		return false
	}

	return *addr == *otherAddr
}

func (addr *ImplicitAccountCreationAddress) Type() AddressType {
	return AddressImplicitAccountCreation
}

func (addr *ImplicitAccountCreationAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *ImplicitAccountCreationAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *ImplicitAccountCreationAddress) Size() int {
	return ImplicitAccountCreationAddressSerializedBytesSize
}

// ImplicitAccountCreationAddressFromPubKey returns the address belonging to the given Ed25519 public key.
func ImplicitAccountCreationAddressFromPubKey(pubKey ed25519.PublicKey) *ImplicitAccountCreationAddress {
	address := blake2b.Sum256(pubKey[:])

	return (*ImplicitAccountCreationAddress)(&address)
}
