package iotago_test

import (
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestTransactionDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTransaction(),
			target: &iotago.Transaction{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_MaxInputsCount(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:      "ok",
			source:    tpkg.RandTransactionWithInputCount(iotago.MaxInputsCount),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name:      "too many inputs",
			source:    tpkg.RandTransactionWithInputCount(iotago.MaxInputsCount + 1),
			target:    &iotago.Transaction{},
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_MaxOutputsCount(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:      "ok",
			source:    tpkg.RandTransactionWithOutputCount(iotago.MaxOutputsCount),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name:      "too many outputs",
			source:    tpkg.RandTransactionWithOutputCount(iotago.MaxOutputsCount + 1),
			target:    &iotago.Transaction{},
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_RefUTXOIndexMax(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.Inputs[iotago.TxEssenceInput]{
				&iotago.UTXOInput{
					TransactionID:          tpkg.RandTransactionID(),
					TransactionOutputIndex: iotago.RefUTXOIndexMax,
				},
			})),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "wrong ref index",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.Inputs[iotago.TxEssenceInput]{
				&iotago.UTXOInput{
					TransactionID:          tpkg.RandTransactionID(),
					TransactionOutputIndex: iotago.RefUTXOIndexMax + 1,
				},
			})),
			target:    &iotago.Transaction{},
			seriErr:   iotago.ErrRefUTXOIndexInvalid,
			deSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
