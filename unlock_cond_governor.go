package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// GovernorAddressUnlockCondition is an UnlockCondition defining the governor identity for an AliasOutput.
type GovernorAddressUnlockCondition struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *GovernorAddressUnlockCondition) Clone() UnlockCondition {
	return &GovernorAddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *GovernorAddressUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *GovernorAddressUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*GovernorAddressUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *GovernorAddressUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionGovernorAddress
}

func (s *GovernorAddressUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionGovernorAddress)) + s.Address.Size()
}
