//nolint:forcetypeassert
package iotago_test

import (
	"crypto/ed25519"
	"testing"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

var (
	benchLargeTx = &iotago.SignedTransaction{
		API: tpkg.TestAPI,
		Transaction: &iotago.Transaction{
			API: tpkg.TestAPI,
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID:     tpkg.TestNetworkID,
				ContextInputs: iotago.TxEssenceContextInputs{},
				Inputs: func() iotago.TxEssenceInputs {
					var inputs iotago.TxEssenceInputs
					for i := 0; i < iotago.MaxInputsCount; i++ {
						inputs = append(inputs, &iotago.UTXOInput{
							TransactionID:          tpkg.Rand36ByteArray(),
							TransactionOutputIndex: 0,
						})
					}

					return inputs
				}(),
				Allotments:   iotago.Allotments{},
				Capabilities: iotago.TransactionCapabilitiesBitMask{},
				Payload:      nil,
			},
			Outputs: func() iotago.TxEssenceOutputs {
				var outputs iotago.TxEssenceOutputs
				for i := 0; i < iotago.MaxOutputsCount; i++ {
					outputs = append(outputs, &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					})
				}

				return outputs
			}(),
		},
		Unlocks: func() iotago.Unlocks {
			var unlocks iotago.Unlocks
			for i := 0; i < iotago.MaxInputsCount; i++ {
				unlocks = append(unlocks, &iotago.SignatureUnlock{
					Signature: tpkg.RandEd25519Signature(),
				})
			}

			return unlocks
		}(),
	}
)

func BenchmarkDeserializationLargeTxPayload(b *testing.B) {
	data, err := tpkg.TestAPI.Encode(benchLargeTx, serix.WithValidation())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("reflection with validation", func(b *testing.B) {
		target := &iotago.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.TestAPI.Decode(data, target, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		target := &iotago.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.TestAPI.Decode(data, target)
		}
	})
}

func BenchmarkDeserializationOneIOTxPayload(b *testing.B) {
	data, err := tpkg.TestAPI.Encode(tpkg.OneInputOutputTransaction(), serix.WithValidation())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("reflection with validation", func(b *testing.B) {
		target := &iotago.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.TestAPI.Decode(data, target, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		target := &iotago.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.TestAPI.Decode(data, target)
		}
	})
}

func BenchmarkSerializationOneIOTxPayload(b *testing.B) {

	b.Run("reflection with validation", func(b *testing.B) {
		txPayload := tpkg.OneInputOutputTransaction()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.TestAPI.Encode(txPayload, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		txPayload := tpkg.OneInputOutputTransaction()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.TestAPI.Encode(txPayload)
		}
	})
}

func BenchmarkSignEd25519OneIOTxEssence(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Transaction.SigningMessage()
	tpkg.Must(err)

	seed := tpkg.RandEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(prvKey, txEssenceData)
	}
}

func BenchmarkVerifyEd25519OneIOTxEssence(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Transaction.SigningMessage()
	tpkg.Must(err)

	seed := tpkg.RandEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	sig := ed25519.Sign(prvKey, txEssenceData)

	pubKey := prvKey.Public().(ed25519.PublicKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hiveEd25519.Verify(pubKey, txEssenceData, sig)
	}
}

func BenchmarkSerializeAndHashBlockWithTransactionPayload(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()

	m := &iotago.Block{
		API: tpkg.TestAPI,
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion: tpkg.TestAPI.Version(),
		},
		Body: &iotago.BasicBlock{
			API:                tpkg.TestAPI,
			StrongParents:      tpkg.SortedRandBlockIDs(2),
			WeakParents:        iotago.BlockIDs{},
			ShallowLikeParents: iotago.BlockIDs{},
			Payload:            txPayload,
		},
	}
	for i := 0; i < b.N; i++ {
		_, _ = m.ID()
	}
}
