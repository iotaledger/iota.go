package builder_test

import (
	"crypto/ed25519"
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestTransactionBuilder(t *testing.T) {
	identityOne := tpkg.RandEd25519PrivateKey()
	inputAddr := iotago.Ed25519AddressFromPubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iotago.AddressSigner
		builder    *builder.TransactionBuilder
		buildErr   error
	}

	tests := []test{
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(rand.Uint64()).
				AddInput(&builder.ToBeSignedUTXOInput{Address: &inputAddr, OutputID: inputUTXO1.ID(), Output: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			return test{
				name:       "ok - 1 input/output",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    bdl,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(rand.Uint64()).
				AddInput(&builder.ToBeSignedUTXOInput{Address: &inputAddr, OutputID: inputUTXO1.ID(), Output: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}).
				AddTaggedDataPayload(&iotago.TaggedData{Tag: []byte("index"), Data: nil})

			return test{
				name:       "ok - with tagged data payload",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    bdl,
			}
		}(),
		func() test {
			bdl := builder.NewTransactionBuilder(rand.Uint64())
			return test{
				name:       "err - no inputs",
				addrSigner: iotago.NewInMemoryAddressSigner(),
				builder:    bdl,
				buildErr:   serializer.ErrArrayValidationMinElementsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			bdl := builder.NewTransactionBuilder(rand.Uint64()).
				AddInput(&builder.ToBeSignedUTXOInput{Address: &inputAddr, OutputID: inputUTXO1.ID(), Output: tpkg.RandBasicOutput(iotago.AddressEd25519)})
			return test{
				name:       "err - no outputs",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    bdl,
				buildErr:   serializer.ErrArrayValidationMinElementsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(rand.Uint64()).
				AddInput(&builder.ToBeSignedUTXOInput{Address: &inputAddr, OutputID: inputUTXO1.ID(), Output: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			// wrong address/keys
			wrongIdentity := tpkg.RandEd25519PrivateKey()
			wrongAddr := iotago.Ed25519AddressFromPubKey(wrongIdentity.Public().(ed25519.PublicKey))
			wrongAddrKeys := iotago.AddressKeys{Address: &wrongAddr, Keys: wrongIdentity}

			return test{
				name:       "err - missing address keys (wrong address)",
				addrSigner: iotago.NewInMemoryAddressSigner(wrongAddrKeys),
				builder:    bdl,
				buildErr:   iotago.ErrAddressKeysNotMapped,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(rand.Uint64()).
				AddInput(&builder.ToBeSignedUTXOInput{Address: &inputAddr, OutputID: inputUTXO1.ID(), Output: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			return test{
				name:       "err - missing address keys (no keys given at all)",
				addrSigner: iotago.NewInMemoryAddressSigner(),
				builder:    bdl,
				buildErr:   iotago.ErrAddressKeysNotMapped,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder.Build(iotago.ZeroRentParas, test.addrSigner)
			if test.buildErr != nil {
				assert.True(t, errors.Is(err, test.buildErr), "wrong error : %s != %s", err, test.buildErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
