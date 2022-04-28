package iotago

import (
	"fmt"
)

// PayloadType denotes a type of payload.
type PayloadType uint32

const (
	// Deprecated payload types
	// PayloadTransactionTIP7 = 0
	// PayloadMilestoneTIP8 = 1
	// PayloadIndexationTIP6 = 2
	// PayloadReceiptTIP17TIP8 = 3.

	// PayloadTreasuryTransaction denotes a TreasuryTransaction.
	PayloadTreasuryTransaction PayloadType = 4
	// PayloadTaggedData denotes a TaggedData payload.
	PayloadTaggedData PayloadType = 5
	// PayloadTransaction denotes a Transaction.
	PayloadTransaction PayloadType = 6
	// PayloadMilestone denotes a Milestone.
	PayloadMilestone PayloadType = 7
)

func (payloadType PayloadType) String() string {
	if int(payloadType) >= len(payloadNames) {
		return fmt.Sprintf("unknown payload type: %d", payloadType)
	}
	return payloadNames[payloadType]
}

var (
	payloadNames = [PayloadMilestone + 1]string{
		"Deprecated-TransactionTIP7",
		"Deprecated-MilestoneTIP8",
		"Deprecated-IndexationTIP6",
		"Deprecated-ReceiptTIP17TIP8",
		"TreasuryTransaction",
		"TaggedData",
		"Transaction",
		"Milestone",
	}
)

// Payload is an object which can be embedded into other objects.
type Payload interface {
	Sizer

	// PayloadType returns the type of the payload.
	PayloadType() PayloadType
}
