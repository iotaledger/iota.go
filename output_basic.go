package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

type (
	BasicOutputUnlockCondition interface{ UnlockCondition }
	BasicOutputFeature         interface{ Feature }
)

// BasicOutputs is a slice of BasicOutput(s).
type BasicOutputs []*BasicOutput

// BasicOutput is an output type which can hold native tokens and features.
type BasicOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64 `serix:"0,mapKey=amount"`
	// The native tokens held by the output.
	NativeTokens NativeTokens `serix:"1,mapKey=nativeTokens,omitempty"`
	// The unlock conditions on this output.
	Conditions UnlockConditions[BasicOutputUnlockCondition] `serix:"2,mapKey=unlockConditions,omitempty"`
	// The features on the output.
	Features Features[BasicOutputFeature] `serix:"3,mapKey=features,omitempty"`
}

// IsSimpleTransfer tells whether this BasicOutput fulfills the criteria of being a simple transfer.
func (e *BasicOutput) IsSimpleTransfer() bool {
	return len(e.FeatureSet()) == 0 && len(e.UnlockConditionSet()) == 1 && len(e.NativeTokens) == 0
}

func (e *BasicOutput) Clone() Output {
	return &BasicOutput{
		Amount:       e.Amount,
		NativeTokens: e.NativeTokens.Clone(),
		Conditions:   e.Conditions.Clone(),
		Features:     e.Features.Clone(),
	}
}

func (e *BasicOutput) UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(e, nil, ident, extParas)
	return ok
}

func (e *BasicOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		e.NativeTokens.VBytes(rentStruct, nil) +
		e.Conditions.VBytes(rentStruct, nil) +
		e.Features.VBytes(rentStruct, nil)
}

func (e *BasicOutput) NativeTokenList() NativeTokens {
	return e.NativeTokens
}

func (e *BasicOutput) FeatureSet() FeatureSet {
	return e.Features.MustSet()
}

func (e *BasicOutput) UnlockConditionSet() UnlockConditionSet {
	return e.Conditions.MustSet()
}
func (e *BasicOutput) Deposit() uint64 {
	return e.Amount
}

func (e *BasicOutput) Ident() Address {
	return e.Conditions.MustSet().Address().Address
}

func (e *BasicOutput) Type() OutputType {
	return OutputBasic
}

func (e *BasicOutput) Size() int {
	return util.NumByteLen(byte(OutputBasic)) +
		util.NumByteLen(e.Amount) +
		e.NativeTokens.Size() +
		e.Conditions.Size() +
		e.Features.Size()
}
