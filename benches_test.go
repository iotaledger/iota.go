package iota_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/iotaledger/iota.go"
)

func BenchmarkDeserializeWithValidationOneIOTxPayload(b *testing.B) {
	data, err := oneInputOutputTransaction().Serialize(iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iota.Transaction{}
	_, err = target.Deserialize(data, iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, iota.DeSeriModePerformValidation)
	}
}

func BenchmarkDeserializeWithoutValidationOneIOTxPayload(b *testing.B) {
	data, err := oneInputOutputTransaction().Serialize(iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iota.Transaction{}
	_, err = target.Deserialize(data, iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, iota.DeSeriModeNoValidation)
	}
}

func BenchmarkSerializeWithValidationOneIOTxPayload(b *testing.B) {
	txPayload := oneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txPayload.Serialize(iota.DeSeriModePerformValidation)
	}
}

func BenchmarkSerializeWithoutValidationOneIOTxPayload(b *testing.B) {
	sigTxPayload := oneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(iota.DeSeriModeNoValidation)
	}
}

func BenchmarkSignEd25519OneIOTxEssence(b *testing.B) {
	txPayload := oneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Essence.Serialize(iota.DeSeriModeNoValidation)
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(prvKey, txEssenceData)
	}
}

func BenchmarkVerifyEd25519OneIOTxEssence(b *testing.B) {
	txPayload := oneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Essence.Serialize(iota.DeSeriModeNoValidation)
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	sig := ed25519.Sign(prvKey, txEssenceData)

	pubKey := prvKey.Public().(ed25519.PublicKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Verify(pubKey, txEssenceData, sig)
	}
}

func BenchmarkSerializeAndHashMessageWithTransactionPayload(b *testing.B) {
	txPayload := oneInputOutputTransaction()
	m := &iota.Message{
		Version: 1,
		Parent1: rand32ByteHash(),
		Parent2: rand32ByteHash(),
		Payload: txPayload,
		Nonce:   0,
	}
	for i := 0; i < b.N; i++ {
		_, _ = m.ID()
	}
}
