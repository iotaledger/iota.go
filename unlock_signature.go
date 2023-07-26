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
	// UnlockType
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(serializer.OneByte)
	if err != nil {
		return 0, err
	}

	workScoreSignature, err := s.Signature.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workScoreBytes.Add(workScoreSignature)
}
