package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

// TimelockUnlockCondition is an unlock condition which puts a time constraint on an output depending
// on the latest confirmed milestone's timestamp T:
//   - the output can only be consumed, if T is bigger than the one defined in the condition.
type TimelockUnlockCondition struct {
	// The slot index until which the timelock applies (inclusive).
	SlotIndex `serix:"0,mapKey=slotIndex,omitempty"`
}

func (s *TimelockUnlockCondition) Clone() UnlockCondition {
	return &TimelockUnlockCondition{
		SlotIndex: s.SlotIndex,
	}
}

func (s *TimelockUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize + serializer.UInt64ByteSize)
}

func (s *TimelockUnlockCondition) WorkScore(workScoreStructure *WorkScoreStructure) WorkScore {
	// TimelockUnlockCondition requires a signature check, but on consumption, not creation.
	return workScoreStructure.Factors.Data.Multiply(s.Size())
}

func (s *TimelockUnlockCondition) Equal(other UnlockCondition) bool {
	otherCond, is := other.(*TimelockUnlockCondition)
	if !is {
		return false
	}

	switch {
	case s.SlotIndex != otherCond.SlotIndex:
		return false
	}

	return true
}

func (s *TimelockUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionTimelock
}

func (s *TimelockUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionTimelock)) +
		len(s.SlotIndex.Bytes())
}
