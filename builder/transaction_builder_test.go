//nolint:scopelint,forcetypeassert
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
	pubKey := prvKey.Public().(ed25519.PublicKey)
	inputAddrEd25519 := iotago.Ed25519AddressFromPubKey(pubKey)
	inputAddrRestricted := iotago.RestrictedAddressWithCapabilities(inputAddrEd25519, iotago.WithAddressCanReceiveAnything())
	inputAddrImplicitAccountCreation := iotago.ImplicitAccountCreationAddressFromPubKey(pubKey)
	signer := iotago.NewInMemoryAddressSignerFromEd25519PrivateKeys(prvKey)

	output := &iotago.BasicOutput{
		Amount: 50,
		UnlockConditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
		},
	}

	type test struct {
		name     string
		builder  *builder.TransactionBuilder
		buildErr error
	}

	tests := []*test{
		// ok - 1 input/output - Ed25519 address
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrEd25519)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: input}).
				AddOutput(output)

			return &test{
				name:    "ok - 1 input/output - Ed25519 address",
				builder: bdl,
			}
		}(),

		// ok - 1 input/output - Restricted address with underlying Ed25519 address
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrRestricted)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrRestricted, InputID: inputUTXO1.OutputID(), Input: input}).
				AddOutput(output)

			return &test{
				name:    "ok - 1 input/output - Restricted address with underlying Ed25519 address",
				builder: bdl,
			}
		}(),

		// ok - 1 input/output - Implicit account creation address
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrImplicitAccountCreation)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrImplicitAccountCreation, InputID: inputUTXO1.OutputID(), Input: input}).
				AddOutput(output)

			return &test{
				name:    "ok - 1 input/output - Implicit account creation address",
				builder: bdl,
			}
		}(),

		// ok - Implicit account creation address with basic input
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			basicInputID := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 1}
			basicOutput := &iotago.BasicOutput{
				Amount:           1000,
				UnlockConditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddrEd25519}},
			}

			input := tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrImplicitAccountCreation)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrImplicitAccountCreation, InputID: inputUTXO1.OutputID(), Input: input}).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: basicInputID.OutputID(), Input: basicOutput}).
				AddOutput(output)

			return &test{
				name:    "ok - Implicit account creation address with basic input",
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
					UnlockConditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddrEd25519}},
				}

				nftOutput = &iotago.NFTOutput{
					Amount:            1000,
					NFTID:             tpkg.Rand32ByteArray(),
					UnlockConditions:  iotago.NFTOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddrEd25519}},
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

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputID1.OutputID(), Input: basicOutput}).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputID2.OutputID(), Input: nftOutput}).
				AddInput(&builder.TxInput{UnlockTarget: nftOutput.ChainID().ToAddress(), InputID: inputID3.OutputID(), Input: accountOwnedByNFT}).
				AddInput(&builder.TxInput{UnlockTarget: accountOwnedByNFT.ChainID().ToAddress(), InputID: inputID4.OutputID(), Input: basicOwnedByAccount}).
				AddOutput(output)

			return &test{
				name:    "ok - mix basic+chain outputs",
				builder: bdl,
			}
		}(),

		// ok - with tagged data payload
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrEd25519)}).
				AddOutput(output).
				AddTaggedDataPayload(&iotago.TaggedData{Tag: []byte("index"), Data: nil})

			return &test{
				name:    "ok - with tagged data payload",
				builder: bdl,
			}
		}(),

		// ok - with context inputs
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrEd25519)}).
				AddOutput(output).
				AddCommitmentInput(&iotago.CommitmentInput{CommitmentID: tpkg.Rand36ByteArray()}).
				AddBlockIssuanceCreditInput(&iotago.BlockIssuanceCreditInput{AccountID: tpkg.RandAccountID()}).
				AddRewardInput(&iotago.RewardInput{Index: 0}, 100)

			return &test{
				name:    "ok - with context inputs",
				builder: bdl,
			}
		}(),

		// ok - allot all mana
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			basicOutput := &iotago.BasicOutput{
				Amount:           1000_000_000,
				UnlockConditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddrEd25519}},
			}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: basicOutput}).
				AddOutput(output).
				AllotAllMana(inputUTXO1.CreationSlot()+6, tpkg.RandAccountID(), 20)

			return &test{
				name:    "ok - allot all mana",
				builder: bdl,
			}
		}(),

		// ok - with mana lock condition
		func() *test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			accountAddr := iotago.AccountAddressFromOutputID(inputUTXO1.OutputID())
			basicOutput := &iotago.BasicOutput{
				Amount: 1000,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: accountAddr},
					&iotago.TimelockUnlockCondition{Slot: inputUTXO1.CreationSlot()},
				}}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrImplicitAccountCreation, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrImplicitAccountCreation)}).
				SetCreationSlot(10).
				AddOutput(basicOutput).
				StoreRemainingManaInOutputAndAllotRemainingAccountBoundMana(inputUTXO1.CreationSlot(), 0)

			return &test{
				name:    "ok - with mana lock condition",
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
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrEd25519)}).
				AddOutput(output)

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
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(iotago.OutputBasic, inputAddrEd25519)}).
				AddOutput(output)

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
