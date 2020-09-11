package iota_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestSignedTransactionBuilder(t *testing.T) {

	identityOne := randEd25519PrivateKey()
	inputAddr := iota.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))

	signerForInputAddr := iota.NewInMemoryAddressSigner(iota.AddressKeys{Address: inputAddr, Keys: identityOne})

	outputAddr1, _ := randEd25519Addr()
	inputUTXO1 := &iota.UTXOInput{TransactionID: randTxHash(), TransactionOutputIndex: 0}

	_, err := iota.NewSignedTransactionBuilder().
		AddInput(&iota.ToBeSignedUTXOInput{Address: inputAddr, Input: inputUTXO1}).
		AddOutput(&iota.SigLockedSingleDeposit{Address: outputAddr1, Amount: 100}).
		Build(signerForInputAddr)
	assert.NoError(t, err)
}
