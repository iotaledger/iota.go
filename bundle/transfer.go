package bundle

import (
	"github.com/iotaledger/giota/pow"
	. "github.com/iotaledger/giota/transaction"
	. "github.com/iotaledger/giota/trinary"
	"time"
)

const (
	// (3^27-1)/2
	UpperBoundAttachmentTimestamp = (3 ^ 27 - 1) / 2
	LowerBoundAttachmentTimestamp = 0
)

type Transfers []Transfer

// Transfer descibes the data/value to transfer to an address.
type Transfer struct {
	Address Hash
	Value   uint64
	Message Trytes
	Tag     Trytes
}

const SignatureMessageFragmentSizeInTrytes = SignatureMessageFragmentTrinarySize / 3

// DoPow computes the nonce field for each transaction so that the last MWM-length trits of the
// transaction hash are all zeroes. Starting from the 0 index transaction, the transactions get chained to
// each other through the trunk transaction hash field. The last transaction in the bundle approves
// the given branch and trunk transactions. This function also initializes the attachment timestamp fields.
func DoPoW(trunkTx, branchTx Trytes, trytes []Trytes, mwm uint64, pow pow.PowFunc) ([]Trytes, error) {
	txs, err := AsTransactionObjects(trytes, nil)
	if err != nil {
		return nil, err
	}
	var prev Trytes
	for i := len(txs) - 1; i >= 0; i-- {
		switch {
		case i == len(txs)-1:
			txs[i].TrunkTransaction = trunkTx
			txs[i].BranchTransaction = branchTx
		default:
			txs[i].TrunkTransaction = prev
			txs[i].BranchTransaction = trunkTx
		}

		txs[i].AttachmentTimestamp = time.Now().UnixNano() / 1000000
		txs[i].AttachmentTimestampLowerBound = LowerBoundAttachmentTimestamp
		txs[i].AttachmentTimestampUpperBound = UpperBoundAttachmentTimestamp

		var err error
		txs[i].Nonce, err = pow(TransactionToTrytes(&txs[i]), int(mwm))
		if err != nil {
			return nil, err
		}

		prev = txs[i].Hash
	}
	powedTxTrytes := TransactionsToTrytes(txs)
	return powedTxTrytes, nil
}
