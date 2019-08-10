package mam

import (
	"strings"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/converter"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
)

// Receiver implementes a receiver for MAM-Messages.
type Receiver struct {
	api  API
	root trinary.Trytes
}

// NewReceiver returns a new receiver.
func NewReceiver(api API, root trinary.Trytes) *Receiver {
	return &Receiver{
		api:  api,
		root: root,
	}
}

// Receive tries to receive all messages from the specified root and returns them.
func (r *Receiver) Receive() ([]string, error) {
	txs, err := r.api.FindTransactionObjects(api.FindTransactionsQuery{Addresses: trinary.Hashes{r.listenAddress()}})
	if err != nil {
		return nil, err
	}

	bundles := map[trinary.Trytes][]transaction.Transaction{}
	for _, tx := range txs {
		if v, ok := bundles[tx.Bundle]; ok {
			bundles[tx.Bundle] = append(v, tx)
		} else {
			bundles[tx.Bundle] = []transaction.Transaction{tx}
		}
	}

	messages := []string{}
	for _, txs := range bundles {
		if len(txs) < int(txs[0].LastIndex+1) {
			continue
		}

		candidateTxs := []transaction.Transaction{}
		for _, tx := range txs {
			if tx.CurrentIndex == 0 {
				candidateTxs = append(candidateTxs, tx)
			}
		}

		for _, candidateTx := range candidateTxs {
			tx := &candidateTx
			message := tx.SignatureMessageFragment
			for tx != nil && tx.CurrentIndex != tx.LastIndex {
				tx = findTxByHash(txs, tx.TrunkTransaction)
				message += tx.SignatureMessageFragment
			}
			if tx.CurrentIndex == tx.LastIndex {
				message, err := r.decodeMessage(message)
				if err != nil {
					return nil, err
				}
				messages = append(messages, message)
			}
		}
	}

	return messages, nil
}

func (r *Receiver) listenAddress() trinary.Hash {
	return r.root
}

func (r *Receiver) decodeMessage(encodedMessage trinary.Trytes) (string, error) {
	messageTrits, err := trinary.TrytesToTrits(encodedMessage)
	if err != nil {
		return "", err
	}
	messageLength := uint64(len(messageTrits))

	rootTrits, err := trinary.TrytesToTrits(r.root)
	if err != nil {
		return "", err
	}

	_, _, messageTrytes, _, err := MAMParse(messageTrits, messageLength, strings.Repeat("9", 81), rootTrits)
	if err != nil {
		return "", err
	}

	message, err := converter.TrytesToASCII(messageTrytes)
	if err != nil {
		return "", err
	}

	return message, nil
}

func findTxByHash(txs transaction.Transactions, hash trinary.Trytes) *transaction.Transaction {
	for _, tx := range txs {
		if tx.Hash == hash {
			return &tx
		}
	}
	return nil
}
