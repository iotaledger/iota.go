package iotago_test

import (
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func BenchmarkDeserializeWithValidationOneIOTxPayload(b *testing.B) {
	data, err := tpkg.OneInputOutputTransaction().Serialize(serializer.DeSeriModeNoValidation, DefZeroRentParas)
	if err != nil {
		b.Fatal(err)
	}

	target := &iotago.Transaction{}
	_, err = target.Deserialize(data, serializer.DeSeriModeNoValidation, DefZeroRentParas)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, serializer.DeSeriModePerformValidation, DefZeroRentParas)
	}
}

func BenchmarkDeserializeWithValidationLargeTxPayload(b *testing.B) {
	origin := &iotago.Transaction{
		Essence: &iotago.TransactionEssence{
			Inputs: func() iotago.Inputs {
				var inputs iotago.Inputs
				for i := 0; i < iotago.MaxInputsCount; i++ {
					inputs = append(inputs, &iotago.UTXOInput{
						TransactionID:          tpkg.Rand32ByteArray(),
						TransactionOutputIndex: 0,
					})
				}
				return inputs
			}(),
			Outputs: func() iotago.Outputs {
				var outputs iotago.Outputs
				for i := 0; i < iotago.MaxOutputsCount; i++ {
					outputs = append(outputs, &iotago.SimpleOutput{
						Address: tpkg.RandEd25519Address(),
						Amount:  100,
					})
				}
				return outputs
			}(),
			Payload: nil,
		},
		UnlockBlocks: func() iotago.UnlockBlocks {
			var unlockBlocks iotago.UnlockBlocks
			for i := 0; i < iotago.MaxInputsCount; i++ {
				unlockBlocks = append(unlockBlocks, &iotago.SignatureUnlockBlock{
					Signature: tpkg.RandEd25519Signature(),
				})
			}
			return unlockBlocks
		}(),
	}

	data, err := origin.Serialize(serializer.DeSeriModePerformValidation|serializer.DeSeriModePerformLexicalOrdering, DefZeroRentParas)
	if err != nil {
		b.Fatal(err)
	}

	target := &iotago.Transaction{}
	_, err = target.Deserialize(data, serializer.DeSeriModeNoValidation, DefZeroRentParas)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, serializer.DeSeriModePerformValidation, DefZeroRentParas)
	}
}

func BenchmarkDeserializeWithoutValidationOneIOTxPayload(b *testing.B) {
	data, err := tpkg.OneInputOutputTransaction().Serialize(serializer.DeSeriModeNoValidation, DefZeroRentParas)
	if err != nil {
		b.Fatal(err)
	}

	target := &iotago.Transaction{}
	_, err = target.Deserialize(data, serializer.DeSeriModeNoValidation, DefZeroRentParas)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, serializer.DeSeriModeNoValidation, DefZeroRentParas)
	}
}

func BenchmarkSerializeWithValidationOneIOTxPayload(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txPayload.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas)
	}
}

func BenchmarkSerializeWithoutValidationOneIOTxPayload(b *testing.B) {
	sigTxPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(serializer.DeSeriModeNoValidation, DefZeroRentParas)
	}
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
		ed25519.Verify(pubKey, txEssenceData, sig)
	}
}

func BenchmarkSerializeAndHashMessageWithTransactionPayload(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()

	m := &iotago.Message{
		Parents: tpkg.SortedRand32BytArray(2),
		Payload: txPayload,
		Nonce:   0,
	}
	for i := 0; i < b.N; i++ {
		_, _ = m.ID()
	}
}
