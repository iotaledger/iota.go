//nolint:dupl
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

func (s *StateControllerAddressUnlockCondition) StorageScore(rentStruct *RentStructure, _ StorageScoreFunc) StorageScore {
	return s.Address.StorageScore(rentStruct, nil)
}

func (s *StateControllerAddressUnlockCondition) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	// StateControllerAddressUnlockCondition does not require a signature check on creation, only consumption.
	return 0, nil
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
