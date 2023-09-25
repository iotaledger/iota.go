package iotago

import (
	"bytes"
	"sort"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// BlockIssuerKeyType defines the type of block issuer key.
type BlockIssuerKeyType byte

const (
	// Ed25519BlockIssuerKey denotes a BlockIssuerKeyEd25519.
	Ed25519BlockIssuerKey BlockIssuerKeyType = iota
	// Ed25519BlockIssuerKeyAddress denotes a BlockIssuerKeyEd25519Address.
	Ed25519BlockIssuerKeyAddress
)

// BlockIssuerKeys are the keys allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKeys []BlockIssuerKey

func (keys BlockIssuerKeys) Sort() {
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i].BlockIssuerKeyBytes(), keys[j].BlockIssuerKeyBytes()) < 0
	})
}

// Size returns the size of the block issuer key when serialized.
func (keys BlockIssuerKeys) Size() int {
	// keys length prefix + size of each key
	size := serializer.OneByte
	for _, key := range keys {
		size += key.Size()
	}

	return size
}

func (keys BlockIssuerKeys) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	// VBFactorIssuerKeys: keys length prefix + each key's vbytes
	vbytes := VBytes(serializer.OneByte)
	for _, key := range keys {
		vbytes += key.VBytes(rentStruct, nil)
	}

	return rentStruct.VBFactorBlockIssuerKey().Multiply(vbytes)
}

// BlockIssuerKey is a key that is allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKey interface {
	Sizer
	NonEphemeralObject

	// BlockIssuerKeyBytes returns a byte slice consisting of the type prefix and the unique identifier of the key.
	BlockIssuerKeyBytes() []byte
	// Type returns the BlockIssuerKeyType.
	Type() BlockIssuerKeyType
}
