package iotago

import (
	"bytes"
	"slices"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
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

func NewBlockIssuerKeys(blockIssuerKey ...BlockIssuerKey) BlockIssuerKeys {
	blockIssuerKeys := make(BlockIssuerKeys, 0, len(blockIssuerKey))
	for _, key := range blockIssuerKey {
		blockIssuerKeys = append(blockIssuerKeys, key)
	}
	blockIssuerKeys.Sort()

	return blockIssuerKeys
}

func (keys BlockIssuerKeys) Clone() BlockIssuerKeys {
	return lo.CloneSlice(keys)
}

// Add adds a new block issuer key if it doesn't exist yet.
func (keys *BlockIssuerKeys) Add(key BlockIssuerKey) {
	for _, k := range *keys {
		if k.Equal(key) {
			// key already exists, don't add it
			return
		}
	}

	// we use the pointer, otherwise the outer slice header is not updated
	*keys = append(*keys, key)
	keys.Sort()
}

// Remove removes a block issuer key in case it exists.
func (keys *BlockIssuerKeys) Remove(key BlockIssuerKey) {
	keysDereferenced := *keys
	for idx, k := range keysDereferenced {
		if k.Equal(key) {
			keysDereferenced = append(keysDereferenced[:idx], keysDereferenced[idx+1:]...)
			keysDereferenced.Sort()

			// we use the pointer, otherwise the outer slice header is not updated
			*keys = keysDereferenced

			return
		}
	}
}

// Has checks if a block issuer key exists.
func (keys BlockIssuerKeys) Has(key BlockIssuerKey) bool {
	for _, k := range keys {
		if k.Equal(key) {
			return true
		}
	}

	return false
}

// Sort sorts the BlockIssuerKeys in place.
func (keys BlockIssuerKeys) Sort() {
	slices.SortFunc(keys, func(x BlockIssuerKey, y BlockIssuerKey) int {
		if x.Type() == y.Type() {
			switch o := x.(type) {
			case *Ed25519AddressBlockIssuerKey:
				//nolint:forcetypeassert
				return o.Compare(y.(*Ed25519AddressBlockIssuerKey))
			case *Ed25519PublicKeyBlockIssuerKey:
				//nolint:forcetypeassert
				return o.Compare(y.(*Ed25519PublicKeyBlockIssuerKey))
			default:
				panic(ierrors.Errorf("unknown block issuer key typ: %T", o))
			}

		}

		return bytes.Compare([]byte{byte(x.Type())}, []byte{byte(y.Type())})
	})
}

func (keys BlockIssuerKeys) Equal(other BlockIssuerKeys) bool {
	if len(keys) != len(other) {
		return false
	}

	for idx, key := range keys {
		if !key.Equal(other[idx]) {
			return false
		}
	}

	return true
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
	constraints.Cloneable[BlockIssuerKey]
	constraints.Equalable[BlockIssuerKey]

	// Bytes returns a byte slice consisting of the type prefix and the unique identifier of the key.
	Bytes() []byte
	// Type returns the BlockIssuerKeyType.
	Type() BlockIssuerKeyType
}
