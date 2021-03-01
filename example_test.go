package iotago_test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/require"
)

func TestSerializedTransactionSize(t *testing.T) {
	sigTxPayload := oneInputOutputTransaction()
	m := &iotago.Message{
		Parents: sortedRand32ByteHashes(2),
		Payload: sigTxPayload,
		Nonce:   0,
	}

	data, err := m.Serialize(iotago.DeSeriModeNoValidation)
	require.NoError(t, err)
	fmt.Printf("length of message cotaining a transaction: %d\n", len(data))
}
