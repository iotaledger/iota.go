package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// ExpirationUnlockCondition is an unlock condition which puts a time constraint on whether the receiver or return identity
// can consume an output depending on the latest confirmed milestone's timestamp T:
//   - only the receiver identity can consume the output, if T is before than the one defined in the condition.
//   - only the return identity can consume the output, if T is at the same time or after the one defined in the condition.
type ExpirationUnlockCondition struct {
	// The identity who is allowed to use the output after the expiration has happened.
	ReturnAddress Address `serix:"0,mapKey=returnAddress"`
	// The slot index at which the expiration happens.
	Slot SlotIndex `serix:"1,mapKey=slot,omitempty"`
}

func (s *ExpirationUnlockCondition) Clone() UnlockCondition {
	return &ExpirationUnlockCondition{
		ReturnAddress: s.ReturnAddress.Clone(),
		Slot:          s.Slot,
	}
}

func (s *ExpirationUnlockCondition) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return s.ReturnAddress.StorageScore(storageScoreStruct, nil)
}

func (s *ExpirationUnlockCondition) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// ExpirationUnlockCondition does not require a signature check on creation, only consumption.
	return 0, nil
}

func (s *ExpirationUnlockCondition) Equal(other UnlockCondition) bool {
	otherCond, is := other.(*ExpirationUnlockCondition)
	if !is {
		return false
	}

	switch {
	case !s.ReturnAddress.Equal(otherCond.ReturnAddress):
		return false
	case s.Slot != otherCond.Slot:
		return false
	}

	return true
}

func (s *ExpirationUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionExpiration
}

func (s *ExpirationUnlockCondition) Size() int {
	// UnlockType + ReturnAddress + SlotIndex
	return serializer.SmallTypeDenotationByteSize + s.ReturnAddress.Size() + SlotIndexLength
}
