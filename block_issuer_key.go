package iotago

import (
	"bytes"
	"slices"

	"github.com/iotaledger/hive.go/ierrors"
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
	BlockIssuerKeyBytes(api API) []byte
	// Type returns the BlockIssuerKeyType.
	Type() BlockIssuerKeyType
	// Equal checks whether other is equal to this BlockIssuerKey.
	Equal(other BlockIssuerKey) bool
}
