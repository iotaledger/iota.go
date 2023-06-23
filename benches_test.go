package iotago_test

import (
	"crypto/ed25519"
	"testing"
	"time"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

var (
	benchLargeTx = &iotago.Transaction{
		Essence: &iotago.TransactionEssence{
			NetworkID: tpkg.TestNetworkID,
			Inputs: func() iotago.TxEssenceInputs {
				var inputs iotago.TxEssenceInputs
				for i := 0; i < iotago.MaxInputsCount; i++ {
					inputs = append(inputs, &iotago.UTXOInput{
						TransactionID:          tpkg.Rand32ByteArray(),
						TransactionOutputIndex: 0,
					})
				}
				return inputs
			}(),
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
			Payload: nil,
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
	data, err := v3API.Encode(benchLargeTx, serix.WithValidation())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("reflection with validation", func(b *testing.B) {
		target := &iotago.Transaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v3API.Decode(data, target, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		target := &iotago.Transaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v3API.Decode(data, target)
		}
	})
}

func BenchmarkDeserializationOneIOTxPayload(b *testing.B) {
	data, err := v3API.Encode(tpkg.OneInputOutputTransaction(), serix.WithValidation())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("reflection with validation", func(b *testing.B) {
		target := &iotago.Transaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v3API.Decode(data, target, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		target := &iotago.Transaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v3API.Decode(data, target)
		}
	})
}

func BenchmarkSerializationOneIOTxPayload(b *testing.B) {

	b.Run("reflection with validation", func(b *testing.B) {
		txPayload := tpkg.OneInputOutputTransaction()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v3API.Encode(txPayload, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		txPayload := tpkg.OneInputOutputTransaction()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v3API.Encode(txPayload)
		}
	})
}

func BenchmarkSignEd25519OneIOTxEssence(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Essence.SigningMessage()
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

	txEssenceData, err := txPayload.Essence.SigningMessage()
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

	slotTimeProvider := iotago.NewTimeProvider(time.Now().Unix(), 10, 10)

	m := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion,
		StrongParents:   tpkg.SortedRandBlockIDs(2),
		Payload:         txPayload,
		Nonce:           0,
	}
	for i := 0; i < b.N; i++ {
		_, _ = m.ID(slotTimeProvider)
	}
}
