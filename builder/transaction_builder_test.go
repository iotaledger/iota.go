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
	identityOne := tpkg.RandEd25519PrivateKey()
	inputAddr := iotago.Ed25519AddressFromPubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iotago.AddressSigner
		builder    *builder.TransactionBuilder
		buildErr   error
	}

	tests := []test{
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandBasicOutput(iotago.AddressEd25519)
			bdl := builder.NewTransactionBuilder(tpkg.TestNetworkID).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.ID(), Input: input}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.BasicOutputUnlockConditions{
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
			var (
				inputID1 = &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
				inputID2 = &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 1}
				inputID3 = &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 4}
				inputID4 = &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 8}
			)

			var (
				basicOutput = &iotago.BasicOutput{
					Amount:     1000,
					Conditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddr}},
				}

				nftOutput = &iotago.NFTOutput{
					Amount:            1000,
					NativeTokens:      nil,
					NFTID:             tpkg.Rand32ByteArray(),
					Conditions:        iotago.NFTOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: inputAddr}},
					Features:          nil,
					ImmutableFeatures: nil,
				}

				accountOwnedByNFT = &iotago.AccountOutput{
					Amount:    1000,
					AccountID: tpkg.Rand32ByteArray(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: nftOutput.Chain().ToAddress()},
						&iotago.GovernorAddressUnlockCondition{Address: nftOutput.Chain().ToAddress()},
					},
				}

				basicOwnedByAccount = &iotago.BasicOutput{
					Amount:     1000,
					Conditions: iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: accountOwnedByNFT.Chain().ToAddress()}},
				}
			)

			bdl := builder.NewTransactionBuilder(tpkg.TestNetworkID).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputID1.ID(), Input: basicOutput}).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputID2.ID(), Input: nftOutput}).
				AddInput(&builder.TxInput{UnlockTarget: nftOutput.Chain().ToAddress(), InputID: inputID3.ID(), Input: accountOwnedByNFT}).
				AddInput(&builder.TxInput{UnlockTarget: accountOwnedByNFT.Chain().ToAddress(), InputID: inputID4.ID(), Input: basicOwnedByAccount}).
				AddOutput(&iotago.BasicOutput{
					Amount: 4000,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			return test{
				name:       "ok - mix basic+chain outputs",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    bdl,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.TestNetworkID).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.ID(), Input: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.BasicOutputUnlockConditions{
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
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.TestNetworkID).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.ID(), Input: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				})

			// wrong address/keys
			wrongIdentity := tpkg.RandEd25519PrivateKey()
			wrongAddr := iotago.Ed25519AddressFromPubKey(wrongIdentity.Public().(ed25519.PublicKey))
			wrongAddrKeys := iotago.AddressKeys{Address: wrongAddr, Keys: wrongIdentity}

			return test{
				name:       "err - missing address keys (wrong address)",
				addrSigner: iotago.NewInMemoryAddressSigner(wrongAddrKeys),
				builder:    bdl,
				buildErr:   iotago.ErrAddressKeysNotMapped,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.TestNetworkID).
				AddInput(&builder.TxInput{UnlockTarget: inputAddr, InputID: inputUTXO1.ID(), Input: tpkg.RandBasicOutput(iotago.AddressEd25519)}).
				AddOutput(&iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.BasicOutputUnlockConditions{
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
			_, err := test.builder.Build(tpkg.TestProtoParams, test.addrSigner)
			if test.buildErr != nil {
				assert.True(t, ierrors.Is(err, test.buildErr), "wrong error : %s != %s", err, test.buildErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
