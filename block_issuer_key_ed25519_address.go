package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// An Ed25519 Address Block Issuer Key.
type BlockIssuerKeyEd25519Address struct {
	Address *Ed25519Address `serix:"0"`
}

// BlockIssuerKeyEd25519AddressFromAddress creates a block issuer key from an Ed25519 address.
func BlockIssuerKeyEd25519AddressFromAddress(address *Ed25519Address) BlockIssuerKeyEd25519Address {
	return BlockIssuerKeyEd25519Address{Address: address}
}

// BlockIssuerKeyBytes returns a byte slice consisting of the type prefix and the raw address.
func (key BlockIssuerKeyEd25519Address) BlockIssuerKeyBytes() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), key))
}

// Type returns the BlockIssuerKeyType.
func (key BlockIssuerKeyEd25519Address) Type() BlockIssuerKeyType {
	return Ed25519BlockIssuerKeyAddress
}

// Size returns the size of the block issuer key when serialized.
func (key BlockIssuerKeyEd25519Address) Size() int {
	return serializer.SmallTypeDenotationByteSize + key.Address.Size()
}

func (key BlockIssuerKeyEd25519Address) VBytes(rentStructure *RentStructure, vbyteFunc VBytesFunc) VBytes {
	return rentStructure.VBFactorData.Multiply(VBytes(serializer.SmallTypeDenotationByteSize)) + key.Address.VBytes(rentStructure, vbyteFunc)
}
