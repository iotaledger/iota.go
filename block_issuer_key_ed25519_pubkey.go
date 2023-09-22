package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// A Ed25519 public key Block Issuer Key.
type BlockIssuerKeyEd25519PublicKey struct {
	PublicKey ed25519.PublicKey `serix:"0"`
}

// BlockIssuerKeyEd25519PublicKeyFromPublicKey creates a block issuer key from an Ed25519 public key.
func BlockIssuerKeyEd25519PublicKeyFromPublicKey(publicKey ed25519.PublicKey) BlockIssuerKeyEd25519PublicKey {
	return BlockIssuerKeyEd25519PublicKey{PublicKey: publicKey}
}

// ToEd25519PublicKey returns the underlying Ed25519 Public Key.
func (key BlockIssuerKeyEd25519PublicKey) ToEd25519PublicKey() ed25519.PublicKey {
	return key.PublicKey
}

// BlockIssuerKeyBytes returns a byte slice consisting of the type prefix and the public key bytes.
func (key BlockIssuerKeyEd25519PublicKey) BlockIssuerKeyBytes() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), key))
}

// Type returns the BlockIssuerKeyType.
func (key BlockIssuerKeyEd25519PublicKey) Type() BlockIssuerKeyType {
	return Ed25519BlockIssuerKeyPublicKey
}

// Size returns the size of the block issuer key when serialized.
func (key BlockIssuerKeyEd25519PublicKey) Size() int {
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize
}

func (key BlockIssuerKeyEd25519PublicKey) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(key.Size()))
}
