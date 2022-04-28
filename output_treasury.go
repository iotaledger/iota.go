package iotago

import (
	"github.com/iotaledger/iota.go/v3/util"
)

// TreasuryOutput is an output which holds the treasury of a network.
type TreasuryOutput struct {
	// The currently residing funds in the treasury.
	Amount uint64 `serix:"0,mapKey=amount"`
}

func (t *TreasuryOutput) NativeTokenList() NativeTokens {
	return nil
}

func (t *TreasuryOutput) UnlockConditionSet() UnlockConditionSet {
	return nil
}

func (t *TreasuryOutput) FeatureSet() FeatureSet {
	return nil
}

func (t *TreasuryOutput) Clone() Output {
	return &TreasuryOutput{Amount: t.Amount}
}

func (t *TreasuryOutput) VBytes(_ *RentStructure, _ VBytesFunc) uint64 {
	return 0
}

func (t *TreasuryOutput) Deposit() uint64 {
	return t.Amount
}

func (t *TreasuryOutput) Type() OutputType {
	return OutputTreasury
}

func (t *TreasuryOutput) Size() int {
	return util.NumByteLen(byte(OutputTreasury)) + util.NumByteLen(t.Amount)
}
