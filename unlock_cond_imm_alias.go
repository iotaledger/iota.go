package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

// ImmutableAccountUnlockCondition is an UnlockCondition defining an account which has to be unlocked.
// Unlike the AddressUnlockCondition, this unlock condition is immutable for an output which contains it,
// meaning it also only applies to ChainOutput(s).
type ImmutableAccountUnlockCondition struct {
	Address *AccountAddress `serix:"0,mapKey=address"`
}

func (s *ImmutableAccountUnlockCondition) Clone() UnlockCondition {
	return &ImmutableAccountUnlockCondition{Address: s.Address.Clone().(*AccountAddress)}
}

func (s *ImmutableAccountUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
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
	return util.NumByteLen(byte(UnlockConditionImmutableAccount)) + s.Address.Size()
}
