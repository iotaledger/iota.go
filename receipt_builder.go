package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// NewReceiptBuilder creates a new ReceiptBuilder.
func NewReceiptBuilder(migratedAt uint32) *ReceiptBuilder {
	return &ReceiptBuilder{
		r: &ReceiptMilestoneOpt{
			MigratedAt:  migratedAt,
			Funds:       MigratedFundsEntries{},
			Transaction: nil,
		},
	}
}

// ReceiptBuilder is used to easily build up a ReceiptMilestoneOpt.
type ReceiptBuilder struct {
	r *ReceiptMilestoneOpt
}

// AddEntry adds the given MigratedFundsEntry to the receipt.
func (rb *ReceiptBuilder) AddEntry(entry *MigratedFundsEntry) *ReceiptBuilder {
	rb.r.Funds = append(rb.r.Funds, entry)
	return rb
}

// AddTreasuryTransaction adds the given TreasuryTransaction to the receipt.
// This function overrides the previously added TreasuryTransaction.
func (rb *ReceiptBuilder) AddTreasuryTransaction(tx *TreasuryTransaction) *ReceiptBuilder {
	rb.r.Transaction = tx
	return rb
}

// Build builds the ReceiptMilestoneOpt.
func (rb *ReceiptBuilder) Build(protoParas *ProtocolParameters) (*ReceiptMilestoneOpt, error) {
	if _, err := rb.r.Serialize(serializer.DeSeriModePerformValidation|serializer.DeSeriModePerformLexicalOrdering, protoParas); err != nil {
		return nil, fmt.Errorf("unable to build receipt: %w", err)
	}
	return rb.r, nil
}
