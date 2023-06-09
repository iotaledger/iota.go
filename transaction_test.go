package iotago_test

import (
	"fmt"
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
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.TxEssenceInputs{
				&iotago.CommitmentInput{
					AccountID:    tpkg.RandAccountID(),
					CommitmentID: iotago.CommitmentID{},
				},
			})),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - BIC",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.TxEssenceInputs{
				&iotago.BICInput{
					AccountID:    tpkg.RandAccountID(),
					CommitmentID: iotago.CommitmentID{},
				},
			})),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - UTXO + Commitment",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.TxEssenceInputs{
				&iotago.UTXOInput{
					TransactionID:          tpkg.RandTransactionID(),
					TransactionOutputIndex: iotago.RefUTXOIndexMax,
				},
				&iotago.CommitmentInput{
					AccountID:    tpkg.RandAccountID(),
					CommitmentID: iotago.CommitmentID{},
				},
			})),
			target:    &iotago.Transaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - UTXO + Commitment + BIC",
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.TxEssenceInputs{
				&iotago.UTXOInput{
					TransactionID:          tpkg.RandTransactionID(),
					TransactionOutputIndex: iotago.RefUTXOIndexMax,
				},
				&iotago.CommitmentInput{
					AccountID:    tpkg.RandAccountID(),
					CommitmentID: iotago.CommitmentID{},
				},
				&iotago.BICInput{
					AccountID:    tpkg.RandAccountID(),
					CommitmentID: iotago.CommitmentID{},
				},
			})),
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
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.TxEssenceInputs{
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
			source: tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.TxEssenceInputs{
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
		AccountID:    tpkg.RandAccountID(),
	}

	commitmentInput2 := &iotago.CommitmentInput{
		CommitmentID: iotago.SlotIdentifierRepresentingData(11, tpkg.RandBytes(32)),
		AccountID:    tpkg.RandAccountID(),
	}
	bicInput1 := &iotago.BICInput{
		CommitmentID: iotago.SlotIdentifierRepresentingData(10, tpkg.RandBytes(32)),
		AccountID:    tpkg.RandAccountID(),
	}
	bicInput2 := &iotago.BICInput{
		CommitmentID: iotago.SlotIdentifierRepresentingData(10, tpkg.RandBytes(32)),
		AccountID:    tpkg.RandAccountID(),
	}

	transaction := tpkg.RandTransactionWithEssence(tpkg.RandTransactionEssenceWithInputs(iotago.TxEssenceInputs{
		utxoInput1,
		commitmentInput1,
		bicInput1,
		commitmentInput2,
		utxoInput2,
		bicInput2,
	}))

	utxoInputs, err := transaction.Inputs()
	require.NoError(t, err)

	commitmentInputs, err := transaction.CommitmentInputs()
	require.NoError(t, err)

	bicInputs, err := transaction.BICInputs()
	require.NoError(t, err)

	fmt.Println(utxoInputs)
	require.Equal(t, 2, len(utxoInputs))
	require.Equal(t, 2, len(commitmentInputs))
	require.Equal(t, 2, len(bicInputs))

	require.Contains(t, utxoInputs, utxoInput1)
	require.Contains(t, utxoInputs, utxoInput2)

	require.Contains(t, commitmentInputs, commitmentInput1)
	require.Contains(t, commitmentInputs, commitmentInput2)

	require.Contains(t, bicInputs, bicInput1)
	require.Contains(t, bicInputs, bicInput2)
}
