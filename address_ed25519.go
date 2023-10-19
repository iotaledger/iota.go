package iotago

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
	// Ed25519AddressBytesLength is the length of an Ed25519 address.
	Ed25519AddressBytesLength = blake2b.Size256
	// Ed25519AddressSerializedBytesSize is the size of a serialized Ed25519 address with its type denoting byte.
	Ed25519AddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + Ed25519AddressBytesLength
)

// Ed25519Address defines an Ed25519 address.
// An Ed25519Address is the Blake2b-256 hash of an Ed25519 public key.
type Ed25519Address [Ed25519AddressBytesLength]byte

func (addr *Ed25519Address) Clone() Address {
	cpy := &Ed25519Address{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *Ed25519Address) StorageScore(_ *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return 0
}

func (addr *Ed25519Address) ID() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr))
}

func (addr *Ed25519Address) Key() string {
	return string(addr.ID())
}

func (addr *Ed25519Address) Unlock(msg []byte, sig Signature) error {
	edSig, isEdSig := sig.(*Ed25519Signature)
	if !isEdSig {
		return ierrors.Wrapf(ErrSignatureAndAddrIncompatible, "can not unlock Ed25519 address with signature of type %s", sig.Type())
	}

	return edSig.Valid(msg, addr)
}

func (addr *Ed25519Address) Equal(other Address) bool {
	otherAddr, is := other.(*Ed25519Address)
	if !is {
		return false
	}

	return *addr == *otherAddr
}

func (addr *Ed25519Address) Type() AddressType {
	return AddressEd25519
}

func (addr *Ed25519Address) Bech32(hrp NetworkPrefix) string {
	return bech32StringBytes(hrp, addr.ID())
}

func (addr *Ed25519Address) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *Ed25519Address) Size() int {
	return Ed25519AddressSerializedBytesSize
}

// Ed25519AddressFromPubKey returns the address belonging to the given Ed25519 public key.
func Ed25519AddressFromPubKey(pubKey ed25519.PublicKey) *Ed25519Address {
	address := blake2b.Sum256(pubKey[:])

	return (*Ed25519Address)(&address)
}
