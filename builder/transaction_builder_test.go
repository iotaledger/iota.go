//nolint:scopelint
package builder_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTransactionBuilder(t *testing.T) {
	prvKey := tpkg.RandEd25519PrivateKey()
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	inputAddr := iotago.Ed25519AddressFromPubKey(prvKey.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: inputAddr, Keys: prvKey}

	type test struct {
		name     string
		builder  *builder.TransactionBuilder
		buildErr error
	}

	tests := []*test{
		// ok - 1 input/output
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandBasicOutput(iotago.AddressEd25519)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, iotago.NewInMemoryAddressSigner(addrKeys)).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.OutputID(), Input: input}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			return &test{
				name:    "ok - 1 input/output",
				builder: bdl,
			}
		}(),

		// ok - mix basic+chain outputs
		func() *test {
			var (
				inputID1 = &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
				inputID2 = &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 1}
				inputID3 = &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 4}
				inputID4 = &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 8}
			)

			var (
				basicOutput = &iotago.BasicOutput{
					Amount:           1000,
					UnlockConditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddr}},
				}

				nftOutput = &iotago.NFTOutput{
					Amount:            1000,
					NFTID:             tpkg.Rand32ByteArray(),
					UnlockConditions:  iotago.NFTOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddr}},
					Features:          nil,
					ImmutableFeatures: nil,
				}

				accountOwnedByNFT = &iotago.AccountOutput{
					Amount:    1000,
					AccountID: tpkg.Rand32ByteArray(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftOutput.ChainID().ToAddress()},
					},
				}

				basicOwnedByAccount = &iotago.BasicOutput{
					Amount:           1000,
					UnlockConditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: accountOwnedByNFT.ChainID().ToAddress()}},
				}
			)

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, iotago.NewInMemoryAddressSigner(addrKeys)).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputID1.OutputID(), Input: basicOutput}).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputID2.OutputID(), Input: nftOutput}).
				AddInput(&builder.TxInput{UnlockTarget: nftOutput.ChainID().ToAddress(), InputID: inputID3.OutputID(), Input: accountOwnedByNFT}).
				AddInput(&builder.TxInput{UnlockTarget: accountOwnedByNFT.ChainID().ToAddress(), InputID: inputID4.OutputID(), Input: basicOwnedByAccount}).
				AddOutput(&iotago.BasicOutput{
					Amount: 4000,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			return &test{
				name:    "ok - mix basic+chain outputs",
				builder: bdl,
			}
		}(),

		// ok - with tagged data payload
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, iotago.NewInMemoryAddressSigner(addrKeys)).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.OutputID(), Input: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}).
				AddTaggedDataPayload(&iotago.TaggedData{Tag: []byte("index"), Data: nil})

			return &test{
				name:    "ok - with tagged data payload",
				builder: bdl,
			}
		}(),

		// err - missing address keys (wrong address)
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			// wrong address/keys
			wrongAddress := tpkg.RandEd25519PrivateKey()
			//nolint:forcetypeassert // we can safely assume that this is a ed25519.PublicKey
			wrongAddr := iotago.Ed25519AddressFromPubKey(wrongAddress.Public().(ed25519.PublicKey))
			wrongAddrKeys := iotago.AddressKeys{Address: wrongAddr, Keys: wrongAddress}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, iotago.NewInMemoryAddressSigner(wrongAddrKeys)).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.OutputID(), Input: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			return &test{
				name:     "err - missing address keys (wrong address)",
				builder:  bdl,
				buildErr: iotago.ErrAddressKeysNotMapped,
			}
		}(),

		// err - missing address keys (no keys given at all)
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, iotago.NewInMemoryAddressSigner()).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.OutputID(), Input: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			return &test{
				name:     "err - missing address keys (no keys given at all)",
				builder:  bdl,
				buildErr: iotago.ErrAddressKeysNotMapped,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder.Build()
			if test.buildErr != nil {
				assert.True(t, ierrors.Is(err, test.buildErr), "wrong error : %s != %s", err, test.buildErr)

				return
			}
			assert.NoError(t, err)
		})
	}
}
