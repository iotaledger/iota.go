package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// ImmutableAccountUnlockCondition is an UnlockCondition defining an account which has to be unlocked.
// Unlike the AddressUnlockCondition, this unlock condition is immutable for an output which contains it,
// meaning it also only applies to ChainOutput(s).
type ImmutableAccountUnlockCondition struct {
	Address *AccountAddress `serix:"0,mapKey=address"`
}

func (s *ImmutableAccountUnlockCondition) Clone() UnlockCondition {
	//nolint:forcetypeassert // we can safely assume that this is an AccountAddress
	return &ImmutableAccountUnlockCondition{Address: s.Address.Clone().(*AccountAddress)}
}

func (s *ImmutableAccountUnlockCondition) StorageScore(rentStruct *RentStructure, _ StorageScoreFunc) StorageScore {
	return s.Address.StorageScore(rentStruct, nil)
}

func (s *ImmutableAccountUnlockCondition) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// ImmutableAccountUnlockCondition does not require a signature check on creation, only consumption.
	return 0, nil
}

func (s *ImmutableAccountUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*ImmutableAccountUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *ImmutableAccountUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionImmutableAccount
}

func (s *ImmutableAccountUnlockCondition) Size() int {
	// UnlockType + Address
	return serializer.SmallTypeDenotationByteSize + s.Address.Size()
}
