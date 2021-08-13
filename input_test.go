package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestInputSelector(t *testing.T) {
	_, err := iotago.InputSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownInputType))
}

func TestInputsValidatorFunc(t *testing.T) {
	type args struct {
		inputs []serializer.Serializable
		funcs  []iotago.InputsValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok addr",
			args{inputs: []serializer.Serializable{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 1,
				},
			}, funcs: []iotago.InputsValidatorFunc{iotago.InputsUTXORefsUniqueValidator()}}, false,
		},
		{
			"addr not unique",
			args{inputs: []serializer.Serializable{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			}, funcs: []iotago.InputsValidatorFunc{iotago.InputsUTXORefsUniqueValidator()}}, true,
		},
		{
			"ok UTXO ref index",
			args{inputs: []serializer.Serializable{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			}, funcs: []iotago.InputsValidatorFunc{iotago.InputsUTXORefIndexBoundsValidator()}}, false,
		},
		{
			"invalid UTXO ref index",
			args{inputs: []serializer.Serializable{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 250,
				},
			}, funcs: []iotago.InputsValidatorFunc{iotago.InputsUTXORefIndexBoundsValidator()}}, true,
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
