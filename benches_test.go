package iotago_test

import (
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
)

func BenchmarkDeserializeWithValidationOneIOTxPayload(b *testing.B) {
	data, err := tpkg.OneInputOutputTransaction().Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iotago.Transaction{}
	_, err = target.Deserialize(data, serializer.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, serializer.DeSeriModePerformValidation)
	}
}

func BenchmarkDeserializeWithoutValidationOneIOTxPayload(b *testing.B) {
	data, err := tpkg.OneInputOutputTransaction().Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iotago.Transaction{}
	_, err = target.Deserialize(data, serializer.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, serializer.DeSeriModeNoValidation)
	}
}

func BenchmarkSerializeWithValidationOneIOTxPayload(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txPayload.Serialize(serializer.DeSeriModePerformValidation)
	}
}

func BenchmarkSerializeWithoutValidationOneIOTxPayload(b *testing.B) {
	sigTxPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(serializer.DeSeriModeNoValidation)
	}
}

func BenchmarkSignEd25519OneIOTxEssence(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Essence.(*iotago.TransactionEssence).SigningMessage()
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

	txEssenceData, err := txPayload.Essence.(*iotago.TransactionEssence).SigningMessage()
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
