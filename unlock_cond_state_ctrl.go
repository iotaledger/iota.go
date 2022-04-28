package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// StateControllerAddressUnlockCondition is an UnlockCondition defining the state controller identity for an AliasOutput.
type StateControllerAddressUnlockCondition struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *StateControllerAddressUnlockCondition) Clone() UnlockCondition {
	return &StateControllerAddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *StateControllerAddressUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *StateControllerAddressUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*StateControllerAddressUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *StateControllerAddressUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionStateControllerAddress
}

func (s *StateControllerAddressUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionStateControllerAddress)) + s.Address.Size()
}
