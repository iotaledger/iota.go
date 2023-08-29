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

// BlockIssuerKeyBytes returns a byte slice consisting of the type prefix and the public key bytes.
func (key BlockIssuerKeyEd25519) BlockIssuerKeyBytes() []byte {
	blockIssuerKeyBytes := make([]byte, 0, key.Size())
	blockIssuerKeyBytes = append(blockIssuerKeyBytes, byte(Ed25519BlockIssuerKey))
	return append(blockIssuerKeyBytes, key.PublicKey[:]...)
}

// Type returns the BlockIssuerKeyType.
func (key BlockIssuerKeyEd25519) Type() BlockIssuerKeyType {
	return Ed25519BlockIssuerKey
}

// Size returns the size of the block issuer key when serialized.
func (key BlockIssuerKeyEd25519) Size() int {
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize
}

func (key BlockIssuerKeyEd25519) VBytes(_ *RentStructure, _ VBytesFunc) VBytes {
	// type prefix + public key size
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize
}
