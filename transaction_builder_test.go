package iotago_test

import (
	"errors"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestTransactionBuilder(t *testing.T) {
	identityOne := tpkg.RandEd25519PrivateKey()
	inputAddr := iotago.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iotago.AddressSigner
		builder    *iotago.TransactionBuilder
		buildErr   error
	}

	tests := []test{
		func() test {
			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - 1 input/output",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
			}
		}(),
		func() test {
			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50}).
				AddIndexationPayload(&iotago.Indexation{Index: []byte("index"), Data: nil})

			return test{
				name:       "ok - with indexation payload",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
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
			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			// wrong address/keys
			wrongIdentity := tpkg.RandEd25519PrivateKey()
			wrongAddr := iotago.AddressFromEd25519PubKey(wrongIdentity.Public().(ed25519.PublicKey))
			wrongAddrKeys := iotago.AddressKeys{Address: &wrongAddr, Keys: wrongIdentity}

			return test{
				name:       "err - missing address keys",
				addrSigner: iotago.NewInMemoryAddressSigner(wrongAddrKeys),
				builder:    builder,
				buildErr:   iotago.ErrAddressKeysNotMapped,
			}
		}(),
		func() test {
			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "err - missing address keys (no keys given at all)",
				addrSigner: iotago.NewInMemoryAddressSigner(),
				builder:    builder,
				buildErr:   iotago.ErrAddressKeysNotMapped,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder.Build(test.addrSigner)
			if test.buildErr != nil {
				assert.True(t, errors.Is(err, test.buildErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}
