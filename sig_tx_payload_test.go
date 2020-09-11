package iota_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestSignedTransactionPayload_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			sigTxPay, sigTxPayData := randSignedTransactionPayload()
			return test{"ok", sigTxPayData, sigTxPay, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iota.SignedTransactionPayload{}
			bytesRead, err := tx.Deserialize(tt.source, iota.DeSeriModePerformValidation)
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

func TestSignedTransactionPayload_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iota.SignedTransactionPayload
		target []byte
	}
	tests := []test{
		func() test {
			sigTxPay, sigTxPayData := randSignedTransactionPayload()
			return test{"ok", sigTxPay, sigTxPayData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestSignedTransactionPayload_SemanticallyValidate(t *testing.T) {
	identityOne := randEd25519PrivateKey()
	inputAddr := iota.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iota.AddressKeys{Address: inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iota.AddressSigner
		builder    *iota.SignedTransactionPayloadBuilder
		inputUTXOs iota.InputToOutputMapping
		buildErr   error
		validErr   error
	}

	tests := []test{
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: randTxHash(), TransactionOutputIndex: 0}

			builder := iota.NewSignedTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleDeposit{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - 1 input/output",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): iota.SigLockedSingleDeposit{Address: inputAddr, Amount: 50},
				},
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			outputAddr2, _ := randEd25519Addr()
			outputAddr3, _ := randEd25519Addr()

			inputUTXO1 := &iota.UTXOInput{TransactionID: randTxHash(), TransactionOutputIndex: 0}
			inputUTXO2 := &iota.UTXOInput{TransactionID: randTxHash(), TransactionOutputIndex: 0}

			builder := iota.NewSignedTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: inputAddr, Input: inputUTXO1}).
				AddInput(&iota.ToBeSignedUTXOInput{Address: inputAddr, Input: inputUTXO2}).
				AddOutput(&iota.SigLockedSingleDeposit{Address: outputAddr1, Amount: 100}).
				AddOutput(&iota.SigLockedSingleDeposit{Address: outputAddr2, Amount: 100}).
				AddOutput(&iota.SigLockedSingleDeposit{Address: outputAddr3, Amount: 100})

			return test{
				name:       "ok - 2 inputs, 3 outputs",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): iota.SigLockedSingleDeposit{Address: inputAddr, Amount: 50},
					inputUTXO2.ID(): iota.SigLockedSingleDeposit{Address: inputAddr, Amount: 250},
				},
			}
		}(),
		func() test {
			builder := iota.NewSignedTransactionBuilder()
			return test{
				name:       "err - no inputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iota.ErrMinInputsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iota.UTXOInput{TransactionID: randTxHash(), TransactionOutputIndex: 0}
			builder := iota.NewSignedTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: inputAddr, Input: inputUTXO1})
			return test{
				name:       "err - no outputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iota.ErrMinOutputsNotReached,
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

			semanticErr := payload.SemanticallyValidate(test.inputUTXOs)
			if test.validErr != nil {
				assert.True(t, errors.Is(semanticErr, test.validErr))
				return
			}
			assert.NoError(t, semanticErr)

			_, err = payload.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
		})
	}

}
