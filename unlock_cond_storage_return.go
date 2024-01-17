package iotago

import (
	"cmp"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// StorageDepositReturnUnlockCondition is an unlock condition which defines
// the amount of tokens which must be sent back to the return identity, when the output in which it occurs in, is consumed.
// If a transaction consumes multiple outputs which have a StorageDepositReturnUnlockCondition, then on the output side at least
// the sum of all occurring StorageDepositReturnUnlockCondition(s) on the input side must be deposited to the designated return identity.
type StorageDepositReturnUnlockCondition struct {
	ReturnAddress Address   `serix:""`
	Amount        BaseToken `serix:""`
}

func (s *StorageDepositReturnUnlockCondition) Clone() UnlockCondition {
	return &StorageDepositReturnUnlockCondition{
		ReturnAddress: s.ReturnAddress.Clone(),
		Amount:        s.Amount,
	}
}

func (s *StorageDepositReturnUnlockCondition) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return s.ReturnAddress.StorageScore(storageScoreStruct, nil)
}

func (s *StorageDepositReturnUnlockCondition) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// StorageDepositReturnUnlockCondition does not require a signature check on creation, only consumption.
	return 0, nil
}

func (s *StorageDepositReturnUnlockCondition) Compare(other UnlockCondition) int {
	return cmp.Compare(s.Type(), other.Type())
}

func (s *StorageDepositReturnUnlockCondition) Equal(other UnlockCondition) bool {
	otherBlock, is := other.(*StorageDepositReturnUnlockCondition)
	if !is {
		return false
	}

	switch {
	case !s.ReturnAddress.Equal(otherBlock.ReturnAddress):
		return false
	case s.Amount != otherBlock.Amount:
		return false
	}

	return true
}

func (s *StorageDepositReturnUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionStorageDepositReturn
}

func (s *StorageDepositReturnUnlockCondition) Size() int {
	// UnlockType + ReturnAddress + BaseToken
	return serializer.SmallTypeDenotationByteSize + s.ReturnAddress.Size() + BaseTokenSize
}
