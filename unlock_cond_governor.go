//nolint:dupl
package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// GovernorAddressUnlockCondition is an UnlockCondition defining the governor identity for an AccountOutput.
type GovernorAddressUnlockCondition struct {
	Address Address `serix:""`
}

func (s *GovernorAddressUnlockCondition) Clone() UnlockCondition {
	return &GovernorAddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *GovernorAddressUnlockCondition) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return s.Address.StorageScore(storageScoreStruct, nil)
}

func (s *GovernorAddressUnlockCondition) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// GovernorAddressUnlockCondition does not require a signature check on creation, only consumption.
	return 0, nil
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
	// UnlockType + Address
	return serializer.SmallTypeDenotationByteSize + s.Address.Size()
}
