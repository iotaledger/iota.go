package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type (
	BasicOutputUnlockCondition  interface{ UnlockCondition }
	BasicOutputFeature          interface{ Feature }
	BasicOutputUnlockConditions = UnlockConditions[BasicOutputUnlockCondition]
	BasicOutputFeatures         = Features[BasicOutputFeature]
)

// BasicOutputs is a slice of BasicOutput(s).
type BasicOutputs []*BasicOutput

// BasicOutput is an output type which can hold native tokens and features.
type BasicOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:""`
	// The stored mana held by the output.
	Mana Mana `serix:""`
	// The unlock conditions on this output.
	UnlockConditions BasicOutputUnlockConditions `serix:",omitempty"`
	// The features on the output.
	Features BasicOutputFeatures `serix:",omitempty"`
}

// IsSimpleTransfer tells whether this BasicOutput fulfills the criteria of being a simple transfer.
func (e *BasicOutput) IsSimpleTransfer() bool {
	return len(e.FeatureSet()) == 0 && len(e.UnlockConditionSet()) == 1
}

func (e *BasicOutput) Clone() Output {
	return &BasicOutput{
		Amount:           e.Amount,
		Mana:             e.Mana,
		UnlockConditions: e.UnlockConditions.Clone(),
		Features:         e.Features.Clone(),
	}
}

func (e *BasicOutput) Equal(other Output) bool {
	otherOutput, isSameType := other.(*BasicOutput)
	if !isSameType {
		return false
	}

	if e.Amount != otherOutput.Amount {
		return false
	}

	if e.Mana != otherOutput.Mana {
		return false
	}

	if !e.UnlockConditions.Equal(otherOutput.UnlockConditions) {
		return false
	}

	if !e.Features.Equal(otherOutput.Features) {
		return false
	}

	return true
}

func (e *BasicOutput) UnlockableBy(ident Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool {
	ok, _ := outputUnlockableBy(e, nil, ident, pastBoundedSlotIndex, futureBoundedSlotIndex)
	return ok
}

func (e *BasicOutput) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStruct.OffsetOutput +
		storageScoreStruct.FactorData().Multiply(StorageScore(e.Size())) +
		e.UnlockConditions.StorageScore(storageScoreStruct, nil) +
		e.Features.StorageScore(storageScoreStruct, nil)
}

func (e *BasicOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreConditions, err := e.UnlockConditions.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreFeatures, err := e.Features.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreConditions.Add(workScoreFeatures)
}

func (e *BasicOutput) FeatureSet() FeatureSet {
	return e.Features.MustSet()
}

func (e *BasicOutput) UnlockConditionSet() UnlockConditionSet {
	return e.UnlockConditions.MustSet()
}

func (e *BasicOutput) BaseTokenAmount() BaseToken {
	return e.Amount
}

func (e *BasicOutput) StoredMana() Mana {
	return e.Mana
}

func (e *BasicOutput) Ident() Address {
	return e.UnlockConditions.MustSet().Address().Address
}

func (e *BasicOutput) Type() OutputType {
	return OutputBasic
}

func (e *BasicOutput) Size() int {
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		ManaSize +
		e.UnlockConditions.Size() +
		e.Features.Size()
}
