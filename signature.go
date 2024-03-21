package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

// Signature is a signature.
type Signature interface {
	serix.Serializable
	serix.Deserializable
	Sizer
	ProcessableObject
	constraints.Cloneable[Signature]

	// MatchesAddress returns true if the Signature matches the given Address.
	MatchesAddress(addr Address) bool

	// SignerUID returns the unique identifier of the signature's signer.
	// This can be used to identify the uniqueness of the signer in the unlocks (e.g. unique public key).
	SignerUID() Identifier

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
