package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
)

func TestInputSelector(t *testing.T) {
	_, err := iotago.InputSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownInputType))
}

func TestInputsValidatorFunc(t *testing.T) {
	type args struct {
		inputs iotago.Inputs
		funcs  []iotago.InputsSyntacticalValidationFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok addr",
			args{inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 1,
				},
			}, funcs: []iotago.InputsSyntacticalValidationFunc{iotago.InputsSyntacticalUnique()}}, false,
		},
		{
			"addr not unique",
			args{inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			}, funcs: []iotago.InputsSyntacticalValidationFunc{iotago.InputsSyntacticalUnique()}}, true,
		},
		{
			"ok UTXO ref index",
			args{inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			}, funcs: []iotago.InputsSyntacticalValidationFunc{iotago.InputsSyntacticalIndicesWithinBounds()}}, false,
		},
		{
			"invalid UTXO ref index",
			args{inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 250,
				},
			}, funcs: []iotago.InputsSyntacticalValidationFunc{iotago.InputsSyntacticalIndicesWithinBounds()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iotago.ValidateInputs(tt.args.inputs, tt.args.funcs...); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
