package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestInputsSyntacticalUnique(t *testing.T) {
	tests := []struct {
		name    string
		inputs  iotago.Inputs
		wantErr error
	}{
		{
			name: "ok",
			inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 1,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - addr not unique",
			inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			},
			wantErr: iotago.ErrInputUTXORefsNotUnique,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.InputsSyntacticalUnique()
			var runErr error
			for index, input := range tt.inputs {
				if err := valFunc(index, input); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestInputsSyntacticalIndicesWithinBounds(t *testing.T) {
	tests := []struct {
		name    string
		inputs  iotago.Inputs
		wantErr error
	}{
		{
			name: "ok",
			inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid UTXO ref index",
			inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 250,
				},
			},
			wantErr: iotago.ErrRefUTXOIndexInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.InputsSyntacticalIndicesWithinBounds()
			var runErr error
			for index, input := range tt.inputs {
				if err := valFunc(index, input); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}
