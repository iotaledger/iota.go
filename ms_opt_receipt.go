package iotago

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrInvalidReceiptMilestoneOpt gets returned when a ReceiptMilestoneOpt is invalid.
	ErrInvalidReceiptMilestoneOpt = errors.New("invalid receipt")
)

const (
	// MinMigratedFundsEntryCount defines the minimum amount of MigratedFundsEntry items within a ReceiptMilestoneOpt.
	MinMigratedFundsEntryCount = 1
	// MaxMigratedFundsEntryCount defines the maximum amount of MigratedFundsEntry items within a ReceiptMilestoneOpt.
	MaxMigratedFundsEntryCount = 127
)

var (
	// ErrReceiptMustContainATreasuryTransaction gets returned if a ReceiptMilestoneOpt does not contain a TreasuryTransaction.
	ErrReceiptMustContainATreasuryTransaction = errors.New("receipt must contain a treasury transaction")
)

// ReceiptMilestoneOpt is a listing of migrated funds.
type ReceiptMilestoneOpt struct {
	// The milestone index at which the funds were migrated in the legacy network.
	MigratedAt MilestoneIndex `serix:"0,mapKey=migratedAt"`
	// Whether this ReceiptMilestoneOpt is the final one for a given migrated at index.
	Final bool `serix:"1,mapKey=final"`
	// The funds which were migrated with this ReceiptMilestoneOpt.
	Funds MigratedFundsEntries `serix:"2,mapKey=funds"`
	// The TreasuryTransaction used to fund the funds.
	Transaction *TreasuryTransaction `serix:"3,optional,mapKey=transaction"`
}

func (r *ReceiptMilestoneOpt) Size() int {
	return serializer.OneByte + serializer.UInt32ByteSize + serializer.OneByte + r.Funds.Size() +
		// payloads have a 4 byte length prefix
		serializer.UInt32ByteSize + r.Transaction.Size()
}

func (r *ReceiptMilestoneOpt) Type() MilestoneOptType {
	return MilestoneOptReceipt
}

func (r *ReceiptMilestoneOpt) Clone() MilestoneOpt {
	return &ReceiptMilestoneOpt{
		MigratedAt:  r.MigratedAt,
		Final:       r.Final,
		Funds:       r.Funds.Clone(),
		Transaction: r.Transaction.Clone(),
	}
}

// SortFunds sorts the funds within the receipt after their serialized binary form in lexical order.
func (r *ReceiptMilestoneOpt) SortFunds() {
	sort.Slice(r.Funds, func(i, j int) bool {
		return bytes.Compare(r.Funds[i].TailTransactionHash[:], r.Funds[j].TailTransactionHash[:]) == -1
	})
}

// Sum returns the sum of all MigratedFundsEntry items within the ReceiptMilestoneOpt.
func (r *ReceiptMilestoneOpt) Sum() uint64 {
	var sum uint64
	for _, item := range r.Funds {
		sum += item.Deposit
	}
	return sum
}

// Treasury returns the TreasuryTransaction within the receipt or nil if none is contained.
// This function panics if the ReceiptMilestoneOpt.Transaction is not nil and not a TreasuryTransaction.
func (r *ReceiptMilestoneOpt) Treasury() *TreasuryTransaction {
	return r.Transaction
}

// ValidateReceipt validates whether given the following receipt:
//   - None of the MigratedFundsEntry objects deposits more than the max supply and deposits at least
//     MinMigratedFundsEntryDeposit tokens.
//   - The sum of all migrated fund entries is not bigger than the total supply.
//   - The previous unspent TreasuryOutput minus the sum of all migrated funds
//     equals the amount of the new TreasuryOutput.
//
// This function panics if the receipt is nil, the receipt does not include any migrated fund entries or
// the given treasury output is nil.
func ValidateReceipt(receipt *ReceiptMilestoneOpt, prevTreasuryOutput *TreasuryOutput, totalSupply uint64) error {
	switch {
	case prevTreasuryOutput == nil:
		panic("given previous treasury output is nil")
	}

	treasuryTransaction := receipt.Treasury()
	if treasuryTransaction == nil {
		return ErrReceiptMustContainATreasuryTransaction
	}

	if receipt.Funds == nil || len(receipt.Funds) == 0 {
		panic("receipt has no migrated funds")
	}

	seenTailTxHashes := make(map[LegacyTailTransactionHash]int)
	var migratedFundsSum uint64
	for fIndex, entry := range receipt.Funds {
		if prevIndex, seen := seenTailTxHashes[entry.TailTransactionHash]; seen {
			return fmt.Errorf("%w: tail transaction hash at index %d occurrs multiple times (previous %d)", ErrInvalidReceiptMilestoneOpt, fIndex, prevIndex)
		}
		seenTailTxHashes[entry.TailTransactionHash] = fIndex

		switch {
		case entry.Deposit < MinMigratedFundsEntryDeposit:
			return fmt.Errorf("%w: migrated fund entry at index %d deposits less than %d", ErrInvalidReceiptMilestoneOpt, fIndex, MinMigratedFundsEntryDeposit)
		case entry.Deposit > totalSupply:
			return fmt.Errorf("%w: migrated fund entry at index %d deposits more than total supply", ErrInvalidReceiptMilestoneOpt, fIndex)
		case entry.Deposit+migratedFundsSum > totalSupply:
			// this can't overflow because the previous case ensures that
			return fmt.Errorf("%w: migrated fund entry at index %d overflows total supply", ErrInvalidReceiptMilestoneOpt, fIndex)
		}

		migratedFundsSum += entry.Deposit
	}

	prevTreasury := prevTreasuryOutput.Amount
	newTreasury := treasuryTransaction.Output.Deposit()
	if prevTreasury-migratedFundsSum != newTreasury {
		return fmt.Errorf("%w: new treasury amount mismatch, prev %d, delta %d (migrated funds), new %d", ErrInvalidReceiptMilestoneOpt, prevTreasury, migratedFundsSum, newTreasury)
	}

	return nil
}
