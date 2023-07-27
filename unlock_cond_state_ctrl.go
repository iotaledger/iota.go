package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// StateControllerAddressUnlockCondition is an UnlockCondition defining the state controller identity for an AccountOutput.
type StateControllerAddressUnlockCondition struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *StateControllerAddressUnlockCondition) Clone() UnlockCondition {
	return &StateControllerAddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *StateControllerAddressUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *StateControllerAddressUnlockCondition) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// StateControllerAddressUnlockCondition does not require a signature check on creation, only consumption.
	return workScoreStructure.DataByte.Multiply(s.Size())
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
	// UnlockType + Address
	return serializer.SmallTypeDenotationByteSize + s.Address.Size()
}
