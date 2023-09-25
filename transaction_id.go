package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

const (
	// TransactionIDLength defines the length of a Transaction ID.
	TransactionIDLength = SlotIndexLength + IdentifierLength
)

var (
	EmptyTransactionID = TransactionID{}

	ErrInvalidTransactionIDLength = ierrors.New("Invalid transactionID length")
)

type TransactionID = SlotIdentifier

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

// TransactionIDFromData returns a new TransactionID for the given data by hashing it with blake2b and appending the creation slot index.
func TransactionIDFromData(creationSlot SlotIndex, data []byte) TransactionID {
	return SlotIdentifierRepresentingData(creationSlot, data)
}
