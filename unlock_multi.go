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

func (u *MultiUnlock) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	var sum WorkScore
	for _, unlock := range u.Unlocks {
		unlockWorkScore, err := unlock.WorkScore(workScoreStructure)
		if err != nil {
			return 0, err
		}

		sum, err = sum.Add(unlockWorkScore)
		if err != nil {
			return 0, err
		}
	}

	return sum, nil
}
