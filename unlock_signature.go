package iotago

import (
	"github.com/iotaledger/iota.go/v3/util"
)

// SignatureUnlock holds a signature which unlocks inputs.
type SignatureUnlock struct {
	// The signature of this unlock.
	Signature Signature `serix:"0,mapKey=signature"`
}

func (s *SignatureUnlock) Type() UnlockType {
	return UnlockSignature
}

func (s *SignatureUnlock) Size() int {
	return util.NumByteLen(byte(UnlockSignature)) + s.Signature.Size()
}
