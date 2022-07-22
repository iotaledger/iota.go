package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// ImmutableAliasUnlockCondition is an UnlockCondition defining an alias which has to be unlocked.
// Unlike the AddressUnlockCondition, this unlock condition is immutable for an output which contains it,
// meaning it also only applies to ChainOutput(s).
type ImmutableAliasUnlockCondition struct {
	Address *AliasAddress `serix:"0,mapKey=address"`
}

func (s *ImmutableAliasUnlockCondition) Clone() UnlockCondition {
	return &ImmutableAliasUnlockCondition{Address: s.Address.Clone().(*AliasAddress)}
}

func (s *ImmutableAliasUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *ImmutableAliasUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*ImmutableAliasUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *ImmutableAliasUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionImmutableAlias
}

func (s *ImmutableAliasUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionImmutableAlias)) + s.Address.Size()
}
