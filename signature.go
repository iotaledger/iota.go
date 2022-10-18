package iotago

import (
	"bytes"
	"fmt"

	_ "golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/core/serix"
)

// Signature is a signature.
type Signature interface {
	serix.Serializable
	serix.Deserializable
	Sizer

	// Type returns the type of the Signature.
	Type() SignatureType
}

// SignatureTypeSet is a set of SignatureType.
type SignatureTypeSet map[SignatureType]struct{}

// SignatureType defines the type of signature.
type SignatureType byte

const (
	// SignatureEd25519 denotes an Ed25519Signature.
	SignatureEd25519 SignatureType = iota
)

func (sigType SignatureType) String() string {
	if int(sigType) >= len(sigNames) {
		return fmt.Sprintf("unknown signature type: %d", sigType)
	}
	return sigNames[sigType]
}

var (
	sigNames = [SignatureEd25519 + 1]string{"Ed25519Signature"}
)

// Signatures is a slice of Signature(s).
type Signatures[T Signature] []T

func (sigs Signatures[T]) Len() int {
	return len(sigs)
}

func (sigs Signatures[T]) Less(i, j int) bool {
	// change this once there are more signature types
	a, b := sigs[i], sigs[j]

	aBytes, _ := _internalAPI.Encode(a)
	bBytes, _ := _internalAPI.Encode(b)

	return bytes.Compare(aBytes, bBytes) < 0
}

func (sigs Signatures[T]) Swap(i, j int) {
	sigs[i], sigs[j] = sigs[j], sigs[i]
}
