//nolint:dupl
package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// AddressUnlockCondition is an UnlockCondition defining an identity which has to be unlocked.
type AddressUnlockCondition struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *AddressUnlockCondition) Clone() UnlockCondition {
	return &AddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *AddressUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return s.Address.VBytes(rentStruct, nil)
}

func (s *AddressUnlockCondition) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	// AddressUnlockCondition does not require a signature check on creation, only consumption.
	return 0, nil
}

func (s *AddressUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*AddressUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *AddressUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionAddress
}

func (s *AddressUnlockCondition) Size() int {
	// UnlockType + Address
	return serializer.SmallTypeDenotationByteSize + s.Address.Size()
}
