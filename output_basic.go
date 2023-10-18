package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type (
	basicOutputUnlockCondition  interface{ UnlockCondition }
	basicOutputFeature          interface{ Feature }
	BasicOutputUnlockConditions = UnlockConditions[basicOutputUnlockCondition]
	BasicOutputFeatures         = Features[basicOutputFeature]
)

// BasicOutputs is a slice of BasicOutput(s).
type BasicOutputs []*BasicOutput

// BasicOutput is an output type which can hold native tokens and features.
type BasicOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:"0,mapKey=amount"`
	// The stored mana held by the output.
	Mana Mana `serix:"1,mapKey=mana"`
	// The unlock conditions on this output.
	Conditions BasicOutputUnlockConditions `serix:"2,mapKey=unlockConditions,omitempty"`
	// The features on the output.
	Features BasicOutputFeatures `serix:"3,mapKey=features,omitempty"`
}

// IsSimpleTransfer tells whether this BasicOutput fulfills the criteria of being a simple transfer.
func (e *BasicOutput) IsSimpleTransfer() bool {
	return len(e.FeatureSet()) == 0 && len(e.UnlockConditionSet()) == 1
}

func (e *BasicOutput) Clone() Output {
	return &BasicOutput{
		Amount:     e.Amount,
		Mana:       e.Mana,
		Conditions: e.Conditions.Clone(),
		Features:   e.Features.Clone(),
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

	if !e.Conditions.Equal(otherOutput.Conditions) {
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

func (e *BasicOutput) StorageScore(rentStruct *RentStructure, _ StorageScoreFunc) StorageScore {
	return offsetOutput(rentStruct) +
		rentStruct.StorageScoreFactorData().Multiply(StorageScore(e.Size())) +
		e.Conditions.StorageScore(rentStruct, nil) +
		e.Features.StorageScore(rentStruct, nil)
}

func (e *BasicOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreConditions, err := e.Conditions.WorkScore(workScoreParameters)
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
	return e.Conditions.MustSet()
}

func (e *BasicOutput) BaseTokenAmount() BaseToken {
	return e.Amount
}

func (e *BasicOutput) StoredMana() Mana {
	return e.Mana
}

func (e *BasicOutput) Ident() Address {
	return e.Conditions.MustSet().Address().Address
}

func (e *BasicOutput) Type() OutputType {
	return OutputBasic
}

func (e *BasicOutput) Size() int {
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		ManaSize +
		e.Conditions.Size() +
		e.Features.Size()
}
