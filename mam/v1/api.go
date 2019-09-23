package mam

import (
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
)

// API defines an interface with a subset of methods of `api.API`.
type API interface {
	PrepareTransfers(seed trinary.Trytes, transfers bundle.Transfers, opts api.PrepareTransfersOptions) ([]trinary.Trytes, error)
	SendTrytes(trytes []trinary.Trytes, depth uint64, mwm uint64, reference ...trinary.Hash) (bundle.Bundle, error)
	FindTransactionObjects(query api.FindTransactionsQuery) (transaction.Transactions, error)
}
