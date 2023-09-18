package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// EmptyUnlock are simply skipped. They are used to maintain correct index relationship between
// addresses and signatures if the signer doesn't know the signature of another signer.
type EmptyUnlock struct{}

func (u *EmptyUnlock) Type() UnlockType {
	return UnlockEmpty
}

func (u *EmptyUnlock) Size() int {
	return serializer.SmallTypeDenotationByteSize
}

func (u *EmptyUnlock) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	return 0, nil
}
