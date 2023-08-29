package iotago

import (
	"bytes"
	"sort"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// BlockIssuerKeyType defines the type of block issuer key.
type BlockIssuerKeyType byte

const (
	// Ed25519BlockIssuerKey denotes a BlockIssuerKeyEd25519.
	Ed25519BlockIssuerKey BlockIssuerKeyType = iota
)

// BlockIssuerKeys are the keys allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKeys []BlockIssuerKey

func (keys BlockIssuerKeys) Sort() {
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i].PublicKeyBytes(), keys[j].PublicKeyBytes()) < 0
	})
}

// Size returns the size of the public key when serialized.
func (keys BlockIssuerKeys) Size() int {
	size := 0
	for _, key := range keys {
		size += key.Size()
	}

	return size
}

func (keys BlockIssuerKeys) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	// VBFactorIssuerKeys: numKeys * (type prefix + pubKeyLength)
	return rentStruct.VBFactorIssuerKeys.Multiply(VBytes(len(keys)) * (serializer.TypeDenotationByteSize + ed25519.PublicKeySize))
}

// BlockIssuerKey is a key that is allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKey interface {
	Sizer

	// PublicKeyBytes returns the Block Issuer Key as a byte slice.
	PublicKeyBytes() []byte
	// Type returns the type of the Block Issuer Key.
	Type() BlockIssuerKeyType
}
