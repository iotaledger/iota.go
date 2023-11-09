package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// TimelockUnlockCondition is an unlock condition which puts a time constraint on an output depending
// on the latest confirmed milestone's timestamp T:
//   - the output can only be consumed, if T is bigger than the one defined in the condition.
type TimelockUnlockCondition struct {
	// The slot index until which the timelock applies (inclusive).
	Slot SlotIndex `serix:",omitempty"`
}

func (s *TimelockUnlockCondition) Clone() UnlockCondition {
	return &TimelockUnlockCondition{
		Slot: s.Slot,
	}
}

func (s *TimelockUnlockCondition) StorageScore(_ *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return 0
}

func (s *TimelockUnlockCondition) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	return 0, nil
}

func (s *TimelockUnlockCondition) Equal(other UnlockCondition) bool {
	otherCond, is := other.(*TimelockUnlockCondition)
	if !is {
		return false
	}

	if s.Slot != otherCond.Slot {
		return false
	}

	return true
}

func (s *TimelockUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionTimelock
}

func (s *TimelockUnlockCondition) Size() int {
	// UnlockType + SlotIndex
	return serializer.SmallTypeDenotationByteSize + SlotIndexLength
}
