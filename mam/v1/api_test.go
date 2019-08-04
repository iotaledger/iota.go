package mam_test

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
)

type fakeAPI struct {
	prepareTransfers       func(trinary.Trytes, bundle.Transfers, api.PrepareTransfersOptions) ([]trinary.Trytes, error)
	sendTrytes             func([]trinary.Trytes, uint64, uint64, ...trinary.Hash) (bundle.Bundle, error)
	findTransactionObjects func(api.FindTransactionsQuery) (transaction.Transactions, error)
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		prepareTransfers: func(_ trinary.Trytes, _ bundle.Transfers, _ api.PrepareTransfersOptions) ([]trinary.Trytes, error) {
			return nil, errors.New("not implemented")
		},
		sendTrytes: func(_ []trinary.Trytes, _ uint64, _ uint64, _ ...trinary.Hash) (bundle.Bundle, error) {
			return nil, errors.New("not implemented")
		},
		findTransactionObjects: func(_ api.FindTransactionsQuery) (transaction.Transactions, error) {
			return nil, errors.New("not implemented")
		},
	}
}

func (f *fakeAPI) PrepareTransfers(seed trinary.Trytes, transfers bundle.Transfers, opts api.PrepareTransfersOptions) ([]trinary.Trytes, error) {
	return f.prepareTransfers(seed, transfers, opts)
}

func (f *fakeAPI) SendTrytes(trytes []trinary.Trytes, depth uint64, mwm uint64, reference ...trinary.Hash) (bundle.Bundle, error) {
	return f.sendTrytes(trytes, depth, mwm, reference...)
}

func (f *fakeAPI) FindTransactionObjects(query api.FindTransactionsQuery) (transaction.Transactions, error) {
	return f.findTransactionObjects(query)
}

func err(returns []interface{}) error {
	if len(returns) == 0 {
		return nil
	}
	last := returns[len(returns)-1]
	if last == nil {
		return nil
	}
	if e, ok := last.(error); ok {
		return e
	}
	return nil
}
