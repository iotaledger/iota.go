package iotago

import (
	"bytes"
	"sort"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// BlockIssuerKeyType defines the type of block issuer key.
type BlockIssuerKeyType byte

const (
	// BlockIssuerKeyEd25519PublicKey denotes a Ed25519PublicKeyBlockIssuerKey.
	BlockIssuerKeyEd25519PublicKey BlockIssuerKeyType = iota
	// BlockIssuerKeyEd25519Address denotes a Ed25519AddressBlockIssuerKey.
	BlockIssuerKeyEd25519Address
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
	vbytes := rentStruct.VBFactorBlockIssuerKey.Multiply(VBytes(serializer.OneByte))
	for _, key := range keys {
		vbytes += key.VBytes(rentStruct, nil)
	}

	return vbytes
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
