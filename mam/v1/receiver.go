package mam

import (
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/converter"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
)

// Receiver implements a receiver for MAM-Messages.
type Receiver struct {
	api     API
	mode    ChannelMode
	sideKey trinary.Trytes
}

// NewReceiver returns a new receiver.
func NewReceiver(api API) *Receiver {
	return &Receiver{
		api:     api,
		mode:    ChannelModePublic,
		sideKey: consts.NullHashTrytes,
	}
}

// SetMode sets the Channel mode.
func (r *Receiver) SetMode(m ChannelMode, sideKey trinary.Trytes) error {
	if m != ChannelModePublic && m != ChannelModePrivate && m != ChannelModeRestricted {
		return ErrUnknownChannelMode
	}
	if m == ChannelModeRestricted {
		if sideKey == "" {
			return ErrNoSideKey
		}
		r.sideKey = sideKey
	}
	r.mode = m
	return nil
}

// Mode returns the Channel mode.
func (r *Receiver) Mode() ChannelMode {
	return r.mode
}

// SideKey returns the Channel's side key.
func (r *Receiver) SideKey() trinary.Trytes {
	return r.sideKey
}

// Receive tries to receive all messages from the specified root and returns them along with the next root.
func (r *Receiver) Receive(root trinary.Trytes) (trinary.Trytes, []string, error) {
	rootTrits, err := trinary.TrytesToTrits(root)
	if err != nil {
		return "", nil, err
	}

	address, err := makeAddress(r.mode, rootTrits, r.sideKey)
	if err != nil {
		return "", nil, err
	}

	txs, err := r.api.FindTransactionObjects(api.FindTransactionsQuery{Addresses: trinary.Hashes{address}})
	if err != nil {
		return "", nil, err
	}

	bundles := map[trinary.Trytes][]transaction.Transaction{}
	for _, tx := range txs {
		if v, ok := bundles[tx.Bundle]; ok {
			bundles[tx.Bundle] = append(v, tx)
		} else {
			bundles[tx.Bundle] = []transaction.Transaction{tx}
		}
	}

	var nextRoot trinary.Trytes
	messages := []string{}
	for _, bundle := range bundles {
		candidateTx := findHeadTx(bundle)
		if candidateTx == nil {
			continue
		}

		tx := candidateTx
		message := tx.SignatureMessageFragment
		for tx != nil && tx.CurrentIndex != tx.LastIndex {
			tx = findTxByHash(bundle, tx.TrunkTransaction)
			if tx != nil {
				message += tx.SignatureMessageFragment
			}
		}
		if tx == nil {
			continue
		}

		if tx.CurrentIndex == tx.LastIndex {
			nr, message, err := r.decodeMessage(rootTrits, message)
			if err != nil {
				return "", nil, err
			}
			messages = append(messages, message)
			nextRoot = nr
		}
	}

	return nextRoot, messages, nil
}

func (r *Receiver) decodeMessage(root trinary.Trits, encodedMessage trinary.Trytes) (trinary.Trytes, string, error) {
	messageTrits, err := trinary.TrytesToTrits(encodedMessage)
	if err != nil {
		return "", "", err
	}
	messageLength := uint64(len(messageTrits))

	_, nextRoot, messageTrytes, _, err := MAMParse(messageTrits, messageLength, r.sideKey, root)
	if err != nil {
		return "", "", err
	}

	message, err := converter.TrytesToASCII(messageTrytes)
	if err != nil {
		return "", "", err
	}

	return nextRoot, message, nil
}

func findHeadTx(txs []transaction.Transaction) *transaction.Transaction {
	for _, tx := range txs {
		if tx.CurrentIndex == 0 {
			return &tx
		}
	}
	return nil
}

func findTxByHash(txs []transaction.Transaction, hash trinary.Trytes) *transaction.Transaction {
	for _, tx := range txs {
		if tx.Hash == hash {
			return &tx
		}
	}
	return nil
}
