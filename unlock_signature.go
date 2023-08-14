package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
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
	// UnlockType + Signature
	return serializer.OneByte + s.Signature.Size()
}

func (s *SignatureUnlock) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	return s.Signature.WorkScore(workScoreStructure)
}
