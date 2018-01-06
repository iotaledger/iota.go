package mamgoiota

import (
	"fmt"
	"sort"
	"time"

	"github.com/iotaledger/mamgoiota/mamutils"

	"github.com/iotaledger/giota"
)

type Transaction struct {
	Message   string
	Value     int64
	Timestamp time.Time
	Recipient string
}

type ApiTransactionsFinder interface {
	FindTransactions(giota.FindTransactionsRequest) ([]giota.Transaction, error)
}

func ReadTransactions(address string, f ApiTransactionsFinder) ([]Transaction, error) {
	iotaAdress, err := giota.ToAddress(address)
	if err != nil {
		return nil, err
	}

	req := giota.FindTransactionsRequest{
		Addresses: []giota.Address{iotaAdress},
	}

	foundTx, err := f.FindTransactions(req)
	if err != nil {
		return nil, err
	}

	sort.Slice(foundTx, func(i, j int) bool {
		return !(foundTx[i].Timestamp.Unix() < foundTx[j].Timestamp.Unix())
	})

	transactions := make([]Transaction, len(foundTx))
	for i, t := range foundTx {
		message, err := mamutils.FromMAMTrytes(t.SignatureMessageFragment)
		if err != nil {
			return nil, err
		}
		transactions[i] = Transaction{
			Message:   message,
			Value:     t.Value,
			Timestamp: t.Timestamp,
			Recipient: string(t.Address),
		}
	}

	return transactions, nil
}

type ApiTransactionsReader interface {
	ReadTransactions([]giota.Trytes) ([]giota.Transaction, error)
}

func ReadTransaction(transactionID string, r ApiTransactionsReader) (Transaction, error) {
	tID, err := giota.ToTrytes(transactionID)
	if err != nil {
		return Transaction{}, err
	}

	txs, err := r.ReadTransactions([]giota.Trytes{tID})
	if len(txs) != 1 {
		return Transaction{}, fmt.Errorf("Requested 1 Transaction but got %d", len(txs))
	}
	if err != nil {
		return Transaction{}, err
	}

	tx := txs[0]
	message, err := mamutils.FromMAMTrytes(tx.SignatureMessageFragment)
	if err != nil {
		return Transaction{}, err
	}
	transaction := Transaction{
		Message:   message,
		Value:     tx.Value,
		Timestamp: tx.Timestamp,
		Recipient: string(tx.Address),
	}

	return transaction, nil
}
