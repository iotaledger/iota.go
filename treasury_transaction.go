package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// TreasuryTransaction represents a transaction which moves funds from the treasury.
type TreasuryTransaction struct {
	// The input of this transaction.
	Input *TreasuryInput `serix:"0,mapKey=input"`
	// The output of this transaction.
	Output *TreasuryOutput `serix:"1,mapKey=output"`
}

func (t *TreasuryTransaction) Size() int {
	return serializer.UInt32ByteSize + t.Input.Size() + t.Output.Size()
}

func (t *TreasuryTransaction) Clone() *TreasuryTransaction {
	return &TreasuryTransaction{
		Input:  t.Input.Clone(),
		Output: t.Output.Clone().(*TreasuryOutput),
	}
}
