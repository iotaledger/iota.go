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

func TestTransactionBuilder(t *testing.T) {
	identityOne := tpkg.RandEd25519PrivateKey()
	inputAddr := iotago.Ed25519AddressFromPubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iotago.AddressSigner
		builder    *iotago.TransactionBuilder
		buildErr   error
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
			}
		}(),
		func() test {
			outputAddr1, _ := tpkg.RandEd25519AddressAndBytes()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr1, Amount: 50}).
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
				addrSigner: iotago.NewInMemoryAddressSigner(),
				builder:    builder,
				buildErr:   serializer.ErrArrayValidationMinElementsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1})
			return test{
				name:       "err - no outputs",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				buildErr:   serializer.ErrArrayValidationMinElementsNotReached,
			}
		}(),
		func() test {
			outputAddr1, _ := tpkg.RandEd25519AddressAndBytes()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr1, Amount: 50})

			// wrong address/keys
			wrongIdentity := tpkg.RandEd25519PrivateKey()
			wrongAddr := iotago.Ed25519AddressFromPubKey(wrongIdentity.Public().(ed25519.PublicKey))
			wrongAddrKeys := iotago.AddressKeys{Address: &wrongAddr, Keys: wrongIdentity}

			return test{
				name:       "err - missing address keys",
				addrSigner: iotago.NewInMemoryAddressSigner(wrongAddrKeys),
				builder:    builder,
				buildErr:   iotago.ErrAddressKeysNotMapped,
			}
		}(),
		func() test {
			outputAddr1, _ := tpkg.RandEd25519AddressAndBytes()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SimpleOutput{Address: outputAddr1, Amount: 50})

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
			_, err := test.builder.Build(DefZeroRentParas, test.addrSigner)
			if test.buildErr != nil {
				assert.True(t, errors.Is(err, test.buildErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}
