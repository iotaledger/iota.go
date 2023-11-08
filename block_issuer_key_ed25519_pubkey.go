package iotago

import (
	"bytes"
	"context"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const Ed25519PublicKeyBlockIssuerKeyLength = serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize

// A Ed25519 public key Block Issuer Key.
type Ed25519PublicKeyBlockIssuerKey struct {
	PublicKey ed25519.PublicKey `serix:""`
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

func Ed25519PublicKeyBlockIssuerKeyFromBytes(bytes []byte) (*Ed25519PublicKeyBlockIssuerKey, error) {
	blockIssuerKey := &Ed25519PublicKeyBlockIssuerKey{}
	_, err := CommonSerixAPI().Decode(context.TODO(), bytes, blockIssuerKey)
	if err != nil {
		return nil, err
	}

	return blockIssuerKey, nil
}

// Bytes returns a byte slice consisting of the type prefix and the public key bytes.
func (key *Ed25519PublicKeyBlockIssuerKey) Bytes() ([]byte, error) {
	return CommonSerixAPI().Encode(context.TODO(), key)
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
	return Ed25519PublicKeyBlockIssuerKeyLength
}

func (key *Ed25519PublicKeyBlockIssuerKey) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStruct.OffsetEd25519BlockIssuerKey()
}
