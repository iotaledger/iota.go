package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestInputSelector(t *testing.T) {
	_, err := iota.InputSelector(100)
	assert.True(t, errors.Is(err, iota.ErrUnknownInputType))
}

func TestInputsValidatorFunc(t *testing.T) {
	type args struct {
		inputs []iota.Serializable
		funcs  []iota.InputsValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok addr",
			args{inputs: []iota.Serializable{
				&iota.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iota.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 1,
				},
			}, funcs: []iota.InputsValidatorFunc{iota.InputsUTXORefsUniqueValidator()}}, false,
		},
		{
			"addr not unique",
			args{inputs: []iota.Serializable{
				&iota.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iota.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			}, funcs: []iota.InputsValidatorFunc{iota.InputsUTXORefsUniqueValidator()}}, true,
		},
		{
			"ok UTXO ref index",
			args{inputs: []iota.Serializable{
				&iota.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			}, funcs: []iota.InputsValidatorFunc{iota.InputsUTXORefIndexBoundsValidator()}}, false,
		},
		{
			"invalid UTXO ref index",
			args{inputs: []iota.Serializable{
				&iota.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 250,
				},
			}, funcs: []iota.InputsValidatorFunc{iota.InputsUTXORefIndexBoundsValidator()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iota.ValidateInputs(tt.args.inputs, tt.args.funcs...); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
