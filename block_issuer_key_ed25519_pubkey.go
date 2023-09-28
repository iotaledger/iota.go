package iotago

import (
	"bytes"
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
func Ed25519PublicKeyBlockIssuerKeyFromPublicKey(publicKey ed25519.PublicKey) *Ed25519PublicKeyBlockIssuerKey {
	return &Ed25519PublicKeyBlockIssuerKey{PublicKey: publicKey}
}

func (key *Ed25519PublicKeyBlockIssuerKey) Clone() BlockIssuerKey {
	return &Ed25519PublicKeyBlockIssuerKey{
		PublicKey: key.PublicKey,
	}
}

// ToEd25519PublicKey returns the underlying Ed25519 Public Key.
func (key *Ed25519PublicKeyBlockIssuerKey) ToEd25519PublicKey() ed25519.PublicKey {
	return key.PublicKey
}

// Bytes returns a byte slice consisting of the type prefix and the public key bytes.
func (key *Ed25519PublicKeyBlockIssuerKey) Bytes() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), key))
}

// Type returns the BlockIssuerKeyType.
func (key *Ed25519PublicKeyBlockIssuerKey) Type() BlockIssuerKeyType {
	return BlockIssuerKeyEd25519PublicKey
}

func (key *Ed25519PublicKeyBlockIssuerKey) Equal(other BlockIssuerKey) bool {
	otherBlockIssuerKey, is := other.(*Ed25519PublicKeyBlockIssuerKey)
	if !is {
		return false
	}

	return bytes.Equal(key.PublicKey[:], otherBlockIssuerKey.PublicKey[:])
}

func (key *Ed25519PublicKeyBlockIssuerKey) Compare(other *Ed25519PublicKeyBlockIssuerKey) int {
	return bytes.Compare(key.PublicKey[:], other.PublicKey[:])
}

// Size returns the size of the block issuer key when serialized.
func (key *Ed25519PublicKeyBlockIssuerKey) Size() int {
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize
}

func (key *Ed25519PublicKeyBlockIssuerKey) StorageScore(rentStruct *RentStructure, _ StorageScoreFunc) StorageScore {
	return rentStruct.StorageScoreOffsetEd25519BlockIssuerKey()
}
