package iotago

import (
	"bytes"
	"context"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// An Ed25519 Address Block Issuer Key.
type Ed25519AddressBlockIssuerKey struct {
	Address *Ed25519Address `serix:"0"`
}

// Ed25519AddressBlockIssuerKeyFromAddress creates a block issuer key from an Ed25519 address.
func Ed25519AddressBlockIssuerKeyFromAddress(address *Ed25519Address) *Ed25519AddressBlockIssuerKey {
	return &Ed25519AddressBlockIssuerKey{Address: address}
}

// BlockIssuerKeyBytes returns a byte slice consisting of the type prefix and the raw address.
func (key *Ed25519AddressBlockIssuerKey) Bytes() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), key))
}

// Type returns the BlockIssuerKeyType.
func (key *Ed25519AddressBlockIssuerKey) Type() BlockIssuerKeyType {
	return BlockIssuerKeyEd25519Address
}

func (key *Ed25519AddressBlockIssuerKey) Equal(other BlockIssuerKey) bool {
	otherBlockIssuerKey, is := other.(*Ed25519AddressBlockIssuerKey)
	if !is {
		return false
	}

	return key.Address.Equal(otherBlockIssuerKey.Address)
}

func (key *Ed25519AddressBlockIssuerKey) Compare(other *Ed25519AddressBlockIssuerKey) int {
	return bytes.Compare(key.Address[:], other.Address[:])
}

// Size returns the size of the block issuer key when serialized.
func (key *Ed25519AddressBlockIssuerKey) Size() int {
	return serializer.SmallTypeDenotationByteSize + key.Address.Size()
}

func (key *Ed25519AddressBlockIssuerKey) VBytes(rentStructure *RentStructure, _ VBytesFunc) VBytes {
	return rentStructure.VBFactorBlockIssuerKey.Multiply(VBytes(key.Size()))
}
