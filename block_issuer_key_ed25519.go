package iotago

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// A Ed25519 Block Issuer Key.
type BlockIssuerKeyEd25519 struct {
	PublicKey ed25519.PublicKey `serix:"0"`
}

// BlockIssuerKeyEd25519FromPublicKey creates a block issuer key from an Ed25519 public key.
func BlockIssuerKeyEd25519FromPublicKey(publicKey ed25519.PublicKey) BlockIssuerKeyEd25519 {
	return BlockIssuerKeyEd25519{PublicKey: publicKey}
}

// ToEd25519PublicKey returns the underlying Ed25519 Public Key.
func (key BlockIssuerKeyEd25519) ToEd25519PublicKey() ed25519.PublicKey {
	return key.PublicKey
}

// PublicKeyBytes returns the public key as a byte slice.
func (key BlockIssuerKeyEd25519) PublicKeyBytes() []byte {
	return key.PublicKey[:]
}

// Type returns the BlockIssuerKeyType.
func (key BlockIssuerKeyEd25519) Type() BlockIssuerKeyType {
	return Ed25519BlockIssuerKey
}

// Size returns the size of the public key when serialized.
func (key BlockIssuerKeyEd25519) Size() int {
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize
}
