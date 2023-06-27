package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

// AddressUnlockCondition is an UnlockCondition defining an identity which has to be unlocked.
type AddressUnlockCondition struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *AddressUnlockCondition) Clone() UnlockCondition {
	return &AddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *AddressUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *AddressUnlockCondition) WorkScore(workScoreStructure *WorkScoreStructure) WorkScore {
	// Address require signature check but this is done on consumption of the output, not creation.
	return workScoreStructure.FactorData.Multiply(s.Size())
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
	return util.NumByteLen(byte(UnlockConditionAddress)) + s.Address.Size()
}
