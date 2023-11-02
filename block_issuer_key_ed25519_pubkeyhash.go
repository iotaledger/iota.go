package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// Ed25519PublicKeyHashBytesLength is the length of an Ed25519 public key hash.
const (
	Ed25519PublicKeyHashBytesLength          = blake2b.Size256
	Ed25519PublicKeyHashBlockIssuerKeyLength = serializer.SmallTypeDenotationByteSize + Ed25519PublicKeyHashBytesLength
)

// An Ed25519 Address Block Issuer Key.
type Ed25519PublicKeyHashBlockIssuerKey struct {
	PublicKeyHash [Ed25519PublicKeyHashBytesLength]byte `serix:"0,mapKey=pubKeyHash"`
}

// Ed25519PublicKeyHashBlockIssuerKeyFromImplicitAccountCreationAddress creates an Ed25519PublicKeyHashBlockIssuerKey from an Ed25519 public key hash.
func Ed25519PublicKeyHashBlockIssuerKeyFromImplicitAccountCreationAddress(address *ImplicitAccountCreationAddress) *Ed25519PublicKeyHashBlockIssuerKey {
	cpy := [Ed25519PublicKeyHashBytesLength]byte{}
	copy(cpy[:], address[:])
	return &Ed25519PublicKeyHashBlockIssuerKey{PublicKeyHash: cpy}
}

// Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey creates an Ed25519PublicKeyHashBlockIssuerKey from an Ed25519 public key.
func Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(pubKey ed25519.PublicKey) *Ed25519PublicKeyHashBlockIssuerKey {
	pubKeyHash := blake2b.Sum256(pubKey[:])
	return &Ed25519PublicKeyHashBlockIssuerKey{
		PublicKeyHash: pubKeyHash,
	}
}

func (key *Ed25519PublicKeyHashBlockIssuerKey) Clone() BlockIssuerKey {
	cpy := [Ed25519PublicKeyHashBytesLength]byte{}
	copy(cpy[:], key.PublicKeyHash[:])
	return &Ed25519PublicKeyHashBlockIssuerKey{
		PublicKeyHash: cpy,
	}
}

func Ed25519PublicKeyHashBlockIssuerKeyFromBytes(bytes []byte) (*Ed25519PublicKeyHashBlockIssuerKey, error) {
	blockIssuerKey := &Ed25519PublicKeyHashBlockIssuerKey{}
	_, err := CommonSerixAPI().Decode(context.TODO(), bytes, blockIssuerKey)
	if err != nil {
		return nil, err
	}

	return blockIssuerKey, nil
}

// Bytes returns a byte slice consisting of the type prefix and the raw address.
func (key *Ed25519PublicKeyHashBlockIssuerKey) Bytes() ([]byte, error) {
	return CommonSerixAPI().Encode(context.TODO(), key)
}

// Type returns the BlockIssuerKeyType.
func (key *Ed25519PublicKeyHashBlockIssuerKey) Type() BlockIssuerKeyType {
	return BlockIssuerKeyPublicKeyHash
}

func (key *Ed25519PublicKeyHashBlockIssuerKey) Equal(other BlockIssuerKey) bool {
	otherBlockIssuerKey, is := other.(*Ed25519PublicKeyHashBlockIssuerKey)
	if !is {
		return false
	}

	return key.PublicKeyHash == otherBlockIssuerKey.PublicKeyHash
}

func (key *Ed25519PublicKeyHashBlockIssuerKey) Compare(other *Ed25519PublicKeyHashBlockIssuerKey) int {
	return bytes.Compare(key.PublicKeyHash[:], other.PublicKeyHash[:])
}

// Size returns the size of the block issuer key when serialized.
func (key *Ed25519PublicKeyHashBlockIssuerKey) Size() int {
	return Ed25519PublicKeyHashBlockIssuerKeyLength
}

func (key *Ed25519PublicKeyHashBlockIssuerKey) StorageScore(storageScoreStructure *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStructure.OffsetEd25519BlockIssuerKey()
}
