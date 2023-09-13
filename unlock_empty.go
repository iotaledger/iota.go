package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

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
