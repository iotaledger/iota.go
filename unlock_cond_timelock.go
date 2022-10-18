package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// TimelockUnlockCondition is an unlock condition which puts a time constraint on an output depending
// on the latest confirmed milestone's timestamp T:
//   - the output can only be consumed, if T is bigger than the one defined in the condition.
type TimelockUnlockCondition struct {
	// The unix time in second resolution until which the timelock applies (inclusive).
	UnixTime uint32 `serix:"0,mapKey=unixTime,omitempty"`
}

func (s *TimelockUnlockCondition) Clone() UnlockCondition {
	return &TimelockUnlockCondition{
		UnixTime: s.UnixTime,
	}
}

func (s *TimelockUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize + serializer.UInt32ByteSize)
}

func (s *TimelockUnlockCondition) Equal(other UnlockCondition) bool {
	otherCond, is := other.(*TimelockUnlockCondition)
	if !is {
		return false
	}

	switch {
	case s.UnixTime != otherCond.UnixTime:
		return false
	}

	return true
}

func (s *TimelockUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionTimelock
}

func (s *TimelockUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionTimelock)) +
		util.NumByteLen(s.UnixTime)
}
