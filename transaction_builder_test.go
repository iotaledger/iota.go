package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestTransactionBuilder(t *testing.T) {
	identityOne := randEd25519PrivateKey()
	inputAddr := iota.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iota.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iota.AddressSigner
		builder    *iota.TransactionBuilder
		buildErr   error
	}

	tests := []test{
		func() test {
			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - 1 input/output",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
			}
		}(),
		func() test {
			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50}).
				AddIndexationPayload(&iota.Indexation{Index: "index", Data: nil})

			return test{
				name:       "ok - with indexation payload",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
			}
		}(),
		func() test {
			builder := iota.NewTransactionBuilder()
			return test{
				name:       "err - no inputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iota.ErrMinInputsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}
			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1})
			return test{
				name:       "err - no outputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iota.ErrMinOutputsNotReached,
			}
		}(),
		func() test {
			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			// wrong address/keys
			wrongIdentity := randEd25519PrivateKey()
			wrongAddr := iota.AddressFromEd25519PubKey(wrongIdentity.Public().(ed25519.PublicKey))
			wrongAddrKeys := iota.AddressKeys{Address: &wrongAddr, Keys: wrongIdentity}

			return test{
				name:       "err - missing address keys",
				addrSigner: iota.NewInMemoryAddressSigner(wrongAddrKeys),
				builder:    builder,
				buildErr:   iota.ErrAddressKeysMissing,
			}
		}(),
		func() test {
			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "err - missing address keys (no keys given at all)",
				addrSigner: iota.NewInMemoryAddressSigner(),
				builder:    builder,
				buildErr:   iota.ErrAddressKeysMissing,
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
