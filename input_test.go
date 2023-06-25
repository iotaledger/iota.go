package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestInputsSyntacticalUnique(t *testing.T) {
	tests := []struct {
		name    string
		inputs  iotago.Inputs[iotago.Input]
		wantErr error
	}{
		{
			name: "ok",
			inputs: iotago.Inputs[iotago.Input]{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 1,
				},
				&iotago.CommitmentInput{
					CommitmentID: tpkg.Rand40ByteArray(),
				},
				&iotago.BICInput{
					AccountID: tpkg.RandAccountID(),
				},
				&iotago.RewardInput{
					Index: 2,
				},
				&iotago.CommitmentInput{
					CommitmentID: tpkg.Rand40ByteArray(),
				},
				&iotago.BICInput{
					AccountID: tpkg.RandAccountID(),
				},
				&iotago.RewardInput{
					Index: 4,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - addr not unique",
			inputs: iotago.Inputs[iotago.Input]{
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
		{
			name: "fail - commitment not unique",
			inputs: iotago.Inputs[iotago.Input]{
				&iotago.CommitmentInput{
					CommitmentID: iotago.CommitmentID{},
				},
				&iotago.CommitmentInput{
					CommitmentID: iotago.CommitmentID{},
				},
			},
			wantErr: iotago.ErrInputCommitmentNotUnique,
		},
		{
			name: "fail - BIC not unique",
			inputs: iotago.Inputs[iotago.Input]{
				&iotago.BICInput{
					AccountID: [32]byte{},
				},
				&iotago.BICInput{
					AccountID: [32]byte{},
				},
			},
			wantErr: iotago.ErrInputBICNotUnique,
		},
		{
			name: "fail - Reward not unique",
			inputs: iotago.Inputs[iotago.Input]{
				&iotago.RewardInput{
					Index: 1,
				},
				&iotago.RewardInput{
					Index: 1,
				},
			},
			wantErr: iotago.ErrInputRewardNotUnique,
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
		inputs  iotago.Inputs[iotago.Input]
		wantErr error
	}{
		{
			name: "ok",
			inputs: iotago.Inputs[iotago.Input]{
				&iotago.UTXOInput{
					TransactionID:          [32]byte{},
					TransactionOutputIndex: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid UTXO ref index",
			inputs: iotago.Inputs[iotago.Input]{
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

func TestInputDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - UTXO",
			source: &iotago.UTXOInput{
				TransactionID:          [32]byte{},
				TransactionOutputIndex: 0,
			},
			target:    &iotago.UTXOInput{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - Commitment",
			source: &iotago.CommitmentInput{
				CommitmentID: iotago.CommitmentID{},
			},
			target:    &iotago.CommitmentInput{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - BIC",
			source: &iotago.BICInput{
				AccountID: tpkg.RandAccountID(),
			},
			target:    &iotago.BICInput{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - Reward",
			source: &iotago.RewardInput{
				Index: 6,
			},
			target:    &iotago.RewardInput{},
			seriErr:   nil,
			deSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
