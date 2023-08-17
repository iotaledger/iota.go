package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// AccountUnlock is an Unlock which references a previous unlock.
type AccountUnlock struct {
	// The other unlock this AccountUnlock references to.
	Reference uint16 `serix:"0,mapKey=reference"`
}

func (r *AccountUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*AccountAddress)

	return ok
}

func (r *AccountUnlock) Chainable() bool {
	return true
}

func (r *AccountUnlock) Ref() uint16 {
	return r.Reference
}

func (r *AccountUnlock) Type() UnlockType {
	return UnlockAccount
}

func (r *AccountUnlock) Size() int {
	// UnlockType + Reference
	return serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
}

func (r *AccountUnlock) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	return 0, nil
}
