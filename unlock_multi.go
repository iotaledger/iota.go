package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// MultiUnlock is an Unlock which holds a list of unlocks for a multi address.
type MultiUnlock struct {
	// The unlocks for this MultiUnlock.
	Unlocks []Unlock `serix:"0,lengthPrefixType=uint8,mapKey=unlocks,minLen=1,maxLen=10"`
}

func (u *MultiUnlock) Type() UnlockType {
	return UnlockMulti
}

func (u *MultiUnlock) Size() int {
	// UnlockType + Unlocks Length
	sum := serializer.SmallTypeDenotationByteSize + serializer.SmallTypeDenotationByteSize

	for _, unlock := range u.Unlocks {
		sum += unlock.Size()
	}

	return sum
}

func (u *MultiUnlock) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	return 0, nil
}
