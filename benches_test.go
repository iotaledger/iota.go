package iota_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/iotaledger/iota.go"
)

func BenchmarkDeserializeWithValidationOneIOSigTxPayload(b *testing.B) {
	data, err := oneInputOutputSignedTransactionPayload().Serialize(iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iota.SignedTransactionPayload{}
	_, err = target.Deserialize(data, iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, iota.DeSeriModePerformValidation)
	}
}

func BenchmarkDeserializeWithoutValidationOneIOSigTxPayload(b *testing.B) {
	data, err := oneInputOutputSignedTransactionPayload().Serialize(iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iota.SignedTransactionPayload{}
	_, err = target.Deserialize(data, iota.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, iota.DeSeriModeNoValidation)
	}
}

func BenchmarkSerializeWithValidationOneIOSigTxPayload(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(iota.DeSeriModePerformValidation)
	}
}

func BenchmarkSerializeWithoutValidationOneIOSigTxPayload(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(iota.DeSeriModeNoValidation)
	}
}

func BenchmarkSignEd25519OneIOUnsignedTx(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()

	unsigTxData, err := sigTxPayload.Transaction.Serialize(iota.DeSeriModeNoValidation)
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(prvKey, unsigTxData)
	}
}

func BenchmarkVerifyEd25519OneIOUnsignedTx(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()

	unsigTxData, err := sigTxPayload.Transaction.Serialize(iota.DeSeriModeNoValidation)
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	sig := ed25519.Sign(prvKey, unsigTxData)

	pubKey := prvKey.Public().(ed25519.PublicKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Verify(pubKey, unsigTxData, sig)
	}
}

func BenchmarkSerializeAndHashMessageWithSignedTransactionPayload(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	m := &iota.Message{
		Version: 1,
		Parent1: randTxHash(),
		Parent2: randTxHash(),
		Payload: sigTxPayload,
		Nonce:   0,
	}
	for i := 0; i < b.N; i++ {
		_, _ = m.Hash()
	}
}
