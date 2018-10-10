package transaction_converter

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/utils"
)

// AsTransactionObject converts transaction trytes of 2673 trytes into a transaction object.
func AsTransactionObject(trytes Trytes, hash ...Hash) (*Transaction, error) {
	if !utils.IsTrytesOfExactLength(trytes, 2673) {
		return nil, ErrInvalidTrytes
	}

	for i := 2279; i < 2295; i++ {
		if trytes[i] != '9' {
			return nil, ErrInvalidTrytes
		}
	}

	trits := MustTrytesToTrits(trytes)

	tx, err := ParseTransaction(trits)
	if err != nil {
		return nil, err
	}

	if len(hash) > 0 {
		tx.Hash = hash[0]
	} else {
		tx.Hash = TransactionHash(tx)
	}
	return tx, nil
}
