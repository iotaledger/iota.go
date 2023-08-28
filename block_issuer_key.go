package iotago

import (
	"bytes"
	"sort"
)

// BlockIssuerKeys are the keys allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKeys []BlockIssuerKey

func (s BlockIssuerKeys) Sort() {
	sort.Slice(s, func(i, j int) bool {
		return bytes.Compare(s[i].PublicKeyBytes(), s[j].PublicKeyBytes()) < 0
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

// TODO: Impl WorkScore func on BlockIssuerKeys.

// BlockIssuerKey is a key that is allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKey interface {
	Sizer

	// PublicKeyBytes returns the Block Issuer Key as a byte slice.
	PublicKeyBytes() []byte
	// Type returns the type of the Block Issuer Key.
	Type() BlockIssuerKeyType
}

// BlockIssuerKeyType defines the type of block issuer key.
type BlockIssuerKeyType byte

const (
	// Ed25519BlockIssuerKey denotes a BlockIssuerKeyEd25519.
	Ed25519BlockIssuerKey BlockIssuerKeyType = iota
)
