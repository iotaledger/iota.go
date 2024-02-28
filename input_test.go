//nolint:scopelint
package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
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
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 1,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - addr not unique",
			inputs: iotago.Inputs[iotago.Input]{
				&iotago.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 0,
				},
				&iotago.UTXOInput{
					TransactionID:          [36]byte{},
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

func TestContextInputsRewardInputMaxIndex(t *testing.T) {
	tests := []struct {
		name    string
		inputs  iotago.ContextInputs[iotago.ContextInput]
		wantErr error
	}{
		{
			name: "ok",
			inputs: iotago.ContextInputs[iotago.ContextInput]{
				&iotago.CommitmentInput{
					CommitmentID: tpkg.Rand36ByteArray(),
				},
				&iotago.BlockIssuanceCreditInput{
					AccountID: tpkg.RandAccountID(),
				},
				&iotago.RewardInput{
					Index: 2,
				},
				&iotago.BlockIssuanceCreditInput{
					AccountID: tpkg.RandAccountID(),
				},
				&iotago.RewardInput{
					Index: 4,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - reward input references index equal to inputs count",
			inputs: iotago.ContextInputs[iotago.ContextInput]{
				&iotago.RewardInput{
					Index: 1,
				},
				&iotago.RewardInput{
					Index: iotago.MaxInputsCount / 2,
				},
			},
			wantErr: iotago.ErrInputRewardIndexExceedsMaxInputsCount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.ContextInputsRewardInputMaxIndex(iotago.MaxInputsCount / 2)
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
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid reference UTXO index",
			inputs: iotago.Inputs[iotago.Input]{
				&iotago.UTXOInput{
					TransactionID:          [36]byte{},
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
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - UTXO",
			Source: &iotago.UTXOInput{
				TransactionID:          [36]byte{},
				TransactionOutputIndex: 0,
			},
			Target:    &iotago.UTXOInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - Commitment",
			Source: &iotago.CommitmentInput{
				CommitmentID: iotago.CommitmentID{},
			},
			Target:    &iotago.CommitmentInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - BIC",
			Source: &iotago.BlockIssuanceCreditInput{
				AccountID: tpkg.RandAccountID(),
			},
			Target:    &iotago.BlockIssuanceCreditInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - Reward",
			Source: &iotago.RewardInput{
				Index: 6,
			},
			Target:    &iotago.RewardInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
