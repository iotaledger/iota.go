package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := tpkg.RandTransaction()
			return test{"ok", txPayloadData, txPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iotago.Transaction{}
			bytesRead, err := tx.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, tx)
		})
	}
}

func TestTransaction_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.Transaction
		target []byte
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := tpkg.RandTransaction()
			return test{"ok", txPayload, txPayloadData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func XTestTransaction_SemanticallyValidate(t *testing.T) {
	identityOne := tpkg.RandEd25519PrivateKey()
	inputAddr := iotago.Ed25519AddressFromPubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iotago.AddressSigner
		builder    *iotago.TransactionBuilder
		inputUTXOs iotago.InputSet
		buildErr   error
		validErr   error
	}

	tests := []test{
		func() test {

			outputAddr1, _ := tpkg.RandEd25519AddressAndBytes()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - 1 input/output",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputSet{
					inputUTXO1.ID(): &iotago.SimpleOutput{Address: &inputAddr, Amount: 50},
				},
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519AddressAndBytes()
			outputAddr2, _ := tpkg.RandEd25519AddressAndBytes()
			outputAddr3, _ := tpkg.RandEd25519AddressAndBytes()
			outputAddr4, _ := tpkg.RandEd25519AddressAndBytes()

			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			inputUTXO2 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO2}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr1, Amount: 20}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr2, Amount: 10}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr3, Amount: 20}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr4, Amount: 1_000_000})

			return test{
				name:       "ok - 2 inputs, 4 outputs",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputSet{
					inputUTXO1.ID(): &iotago.SimpleOutput{Address: &inputAddr, Amount: 50},
					inputUTXO2.ID(): &iotago.SimpleOutput{Address: &inputAddr, Amount: 1_000_000},
				},
			}
		}(),
		func() test {
			builder := iotago.NewTransactionBuilder()
			return test{
				name:       "err - no inputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iotago.ErrMinInputsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1})
			return test{
				name:       "err - no outputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iotago.ErrMinOutputsNotReached,
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519AddressAndBytes()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr1, Amount: 100})

			return test{
				name:       "err - input output sum mismatch",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				validErr:   iotago.ErrInputOutputSumMismatch,
				inputUTXOs: iotago.InputSet{
					inputUTXO1.ID(): &iotago.SimpleOutput{Address: &inputAddr, Amount: 50},
				},
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			payload, err := test.builder.Build(test.addrSigner)
			if test.buildErr != nil {
				assert.True(t, errors.Is(err, test.buildErr))
				return
			}
			assert.NoError(t, err)

			semanticErr := payload.SemanticallyValidate(nil, test.inputUTXOs)
			if test.validErr != nil {
				assert.True(t, errors.Is(semanticErr, test.validErr))
				return
			}
			assert.NoError(t, semanticErr)

			_, err = payload.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
		})
	}

}
