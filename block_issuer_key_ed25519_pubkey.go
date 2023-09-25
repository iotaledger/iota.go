package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// A Ed25519 public key Block Issuer Key.
type Ed25519PublicKeyBlockIssuerKey struct {
	PublicKey ed25519.PublicKey `serix:"0"`
}

// Ed25519PublicKeyBlockIssuerKeyFromPublicKey creates a block issuer key from an Ed25519 public key.
func Ed25519PublicKeyBlockIssuerKeyFromPublicKey(publicKey ed25519.PublicKey) Ed25519PublicKeyBlockIssuerKey {
	return Ed25519PublicKeyBlockIssuerKey{PublicKey: publicKey}
}

// ToEd25519PublicKey returns the underlying Ed25519 Public Key.
func (key Ed25519PublicKeyBlockIssuerKey) ToEd25519PublicKey() ed25519.PublicKey {
	return key.PublicKey
}

// BlockIssuerKeyBytes returns a byte slice consisting of the type prefix and the public key bytes.
func (key Ed25519PublicKeyBlockIssuerKey) BlockIssuerKeyBytes() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), key))
}

// Type returns the BlockIssuerKeyType.
func (key Ed25519PublicKeyBlockIssuerKey) Type() BlockIssuerKeyType {
	return BlockIssuerKeyEd25519PublicKey
}

// Size returns the size of the block issuer key when serialized.
func (key Ed25519PublicKeyBlockIssuerKey) Size() int {
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize
}

func (key Ed25519PublicKeyBlockIssuerKey) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorBlockIssuerKey().Multiply(VBytes(key.Size()))
}
