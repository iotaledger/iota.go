package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTransactionDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - UTXO",
			source: tpkg.RandTransaction(),
			target: &iotago.Transaction{},
		},
		{
			name: "ok -  Commitment",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(
				tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
					&iotago.CommitmentInput{
						CommitmentID: iotago.CommitmentID{},
					},
				}),
			)),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - BIC",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(
				tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
					&iotago.BlockIssuanceCreditInput{
						AccountID: tpkg.RandAccountID(),
					},
				}),
			)),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - Commitment + BIC",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(
				tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
					&iotago.CommitmentInput{
						CommitmentID: iotago.CommitmentID{},
					},
					&iotago.BlockIssuanceCreditInput{
						AccountID: tpkg.RandAccountID(),
					},
				}),
			)),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_MaxInputsCount(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(
				tpkg.WithUTXOInputCount(iotago.MaxInputsCount),
				tpkg.WithBICInputCount(iotago.MaxContextInputsCount/2),
				tpkg.WithRewardInputCount(iotago.MaxContextInputsCount/2-1),
				tpkg.WithCommitmentInput(),
			)),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "too many inputs",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(
				tpkg.WithUTXOInputCount(iotago.MaxInputsCount + 1),
			)),
			target:    &iotago.Transaction{},
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: nil,
		},
		{
			name: "too many context inputs",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(
				tpkg.WithBICInputCount(iotago.MaxContextInputsCount/2),
				tpkg.WithRewardInputCount(iotago.MaxContextInputsCount/2),
				tpkg.WithCommitmentInput(),
			)),
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

func TestTransactionDeSerialize_MaxAllotmentsCount(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:      "ok",
			source:    tpkg.RandTransactionWithAllotmentCount(iotago.MaxAllotmentCount),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name:      "too many outputs",
			source:    tpkg.RandTransactionWithAllotmentCount(iotago.MaxAllotmentCount + 1),
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
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(tpkg.WithInputs(iotago.TxEssenceInputs{
				&iotago.UTXOInput{
					TransactionID:          tpkg.RandTransactionID(),
					TransactionOutputIndex: iotago.RefUTXOIndexMax,
				},
			}))),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "wrong ref index",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(tpkg.WithInputs(iotago.TxEssenceInputs{
				&iotago.UTXOInput{
					TransactionID:          tpkg.RandTransactionID(),
					TransactionOutputIndex: iotago.RefUTXOIndexMax + 1,
				},
			}))),
			target:    &iotago.Transaction{},
			seriErr:   iotago.ErrRefUTXOIndexInvalid,
			deSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransaction_InputTypes(t *testing.T) {
	utxoInput1 := &iotago.UTXOInput{
		TransactionID:          tpkg.RandTransactionID(),
		TransactionOutputIndex: 13,
	}

	utxoInput2 := &iotago.UTXOInput{
		TransactionID:          tpkg.RandTransactionID(),
		TransactionOutputIndex: 11,
	}

	commitmentInput1 := &iotago.CommitmentInput{
		CommitmentID: iotago.SlotIdentifierRepresentingData(10, tpkg.RandBytes(32)),
	}

	bicInput1 := &iotago.BlockIssuanceCreditInput{
		AccountID: tpkg.RandAccountID(),
	}
	bicInput2 := &iotago.BlockIssuanceCreditInput{
		AccountID: tpkg.RandAccountID(),
	}

	rewardInput1 := &iotago.RewardInput{
		Index: 3,
	}
	rewardInput2 := &iotago.RewardInput{
		Index: 2,
	}

	transaction := tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithOptions(
		tpkg.WithInputs(iotago.TxEssenceInputs{
			utxoInput1,
			utxoInput2,
		}),
		tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
			commitmentInput1,
			bicInput1,
			bicInput2,
			rewardInput1,
			rewardInput2,
		}),
	))

	utxoInputs, err := transaction.Inputs()
	require.NoError(t, err)

	commitmentInput := transaction.CommitmentInput()
	require.NotNil(t, commitmentInput)

	bicInputs, err := transaction.BICInputs()
	require.NoError(t, err)

	rewardInputs, err := transaction.RewardInputs()
	require.NoError(t, err)

	require.Equal(t, 2, len(utxoInputs))
	require.Equal(t, 2, len(bicInputs))
	require.Equal(t, 2, len(rewardInputs))

	require.Contains(t, utxoInputs, utxoInput1)
	require.Contains(t, utxoInputs, utxoInput2)

	require.Equal(t, commitmentInput, commitmentInput1)

	require.Contains(t, bicInputs, bicInput1)
	require.Contains(t, bicInputs, bicInput2)

	require.Contains(t, rewardInputs, rewardInput1)
	require.Contains(t, rewardInputs, rewardInput2)
}
