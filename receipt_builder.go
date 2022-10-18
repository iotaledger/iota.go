package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/serix"
)

// NewReceiptBuilder creates a new ReceiptBuilder.
func NewReceiptBuilder(migratedAt MilestoneIndex) *ReceiptBuilder {
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
func (rb *ReceiptBuilder) Build() (*ReceiptMilestoneOpt, error) {
	if _, err := internalEncode(rb.r, serix.WithValidation()); err != nil {
		return nil, fmt.Errorf("unable to build receipt: %w", err)
	}
	return rb.r, nil
}
