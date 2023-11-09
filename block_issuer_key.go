package iotago

import (
	"bytes"
	"context"
	"io"
	"slices"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/stream"
)

// BlockIssuerKeyType defines the type of block issuer key.
type BlockIssuerKeyType byte

const (
	// BlockIssuerKeyEd25519PublicKey denotes a Ed25519PublicKeyBlockIssuerKey.
	BlockIssuerKeyEd25519PublicKey BlockIssuerKeyType = iota
	// BlockIssuerKeyPublicKeyHash denotes a Ed25519PublicKeyHashBlockIssuerKey.
	BlockIssuerKeyPublicKeyHash
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
			case *Ed25519PublicKeyHashBlockIssuerKey:
				//nolint:forcetypeassert
				return o.Compare(y.(*Ed25519PublicKeyHashBlockIssuerKey))
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

func (keys BlockIssuerKeys) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	var storageScore StorageScore
	for _, key := range keys {
		storageScore += key.StorageScore(storageScoreStruct, nil)
	}

	return storageScore
}

func (keys BlockIssuerKeys) Bytes() ([]byte, error) {
	return CommonSerixAPI().Encode(context.TODO(), keys)
}

// BlockIssuerKey is a key that is allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKey interface {
	Sizer
	NonEphemeralObject
	constraints.Cloneable[BlockIssuerKey]
	constraints.Equalable[BlockIssuerKey]
	serializer.Byter

	// Type returns the BlockIssuerKeyType.
	Type() BlockIssuerKeyType
}

func BlockIssuerKeysFromReader(reader io.ReadSeeker) (BlockIssuerKeys, error) {
	b := make(BlockIssuerKeys, 0)
	// serix.LengthPrefixTypeAsByte is used to read the length prefix of the array
	if err := stream.ReadCollection(reader, serializer.SeriLengthPrefixTypeAsByte, func(i int) error {
		blockIssuerKey, err := BlockIssuerKeyFromReader(reader)
		if err != nil {
			return ierrors.Wrapf(err, "unable to read block issuer key %d", i)
		}

		b = append(b, blockIssuerKey)

		return nil
	}); err != nil {
		return nil, ierrors.Wrap(err, "unable to read block issuer keys")
	}

	return b, nil
}

func BlockIssuerKeyFromReader(reader io.ReadSeeker) (BlockIssuerKey, error) {
	blockIssuerKeyType, err := stream.PeekSize(reader, serializer.SeriLengthPrefixTypeAsByte)
	if err != nil {
		return nil, ierrors.Wrap(err, "unable to read block issuer key type")
	}

	switch BlockIssuerKeyType(blockIssuerKeyType) {
	case BlockIssuerKeyEd25519PublicKey:
		readBytes, err := stream.ReadBytes(reader, Ed25519PublicKeyBlockIssuerKeyLength)
		if err != nil {
			return nil, ierrors.Wrap(err, "unable to read block issuer key bytes")
		}

		return Ed25519PublicKeyBlockIssuerKeyFromBytes(readBytes)
	case BlockIssuerKeyPublicKeyHash:
		readBytes, err := stream.ReadBytes(reader, Ed25519PublicKeyHashBlockIssuerKeyLength)
		if err != nil {
			return nil, ierrors.Wrap(err, "unable to read block issuer key bytes")
		}

		return Ed25519PublicKeyHashBlockIssuerKeyFromBytes(readBytes)
	default:
		return nil, ierrors.Errorf("unsupported block issuer key type: %d", blockIssuerKeyType)
	}
}

func BlockIssuerKeyFromBytes(bytes []byte) (BlockIssuerKey, error) {
	var blockIssuerKey BlockIssuerKey

	if _, err := CommonSerixAPI().Decode(context.TODO(), bytes, blockIssuerKey); err != nil {
		return nil, ierrors.Wrap(err, "unable to decode block issuer key")
	}

	return blockIssuerKey, nil
}
