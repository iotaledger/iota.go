package bundle

import (
	"github.com/iotaledger/giota/pow"
	. "github.com/iotaledger/giota/signing"
	. "github.com/iotaledger/giota/transaction"
	. "github.com/iotaledger/giota/trinary"
	"math"
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
	Address AddressHash
	Value   int64
	Message Trytes
	Tag     Trytes
}

const SignatureMessageFragmentSizeInTrytes = SignatureMessageFragmentTrinarySize / 3

// CreateBundle translates the transfer objects into a bundle consisting of all output transactions.
// If a transfer object's message exceeds the signature message fragment size (2187 trytes),
// additional transactions are added to the bundle to accustom the signature fragments.
func CreateBundle(trs Transfers) (bundle Bundle, fragments []Trytes, total int64) {
	for _, tr := range trs {
		numSignatures := 1

		// if the message is longer than 2187 trytes, increase the amount of transactions for the entry
		if len(tr.Message) > SignatureMessageFragmentSizeInTrytes {
			// get total length, message / signature message fragment (2187 trytes)
			fragementsLength := int(math.Floor(float64(len(tr.Message)) / SignatureMessageFragmentSizeInTrytes))
			numSignatures += fragementsLength

			// copy out every fragment
			for k := 0; k < fragementsLength; k++ {
				var fragment Trytes
				switch {
				// remainder
				case k == fragementsLength-1:
					fragment = tr.Message[k*SignatureMessageFragmentSizeInTrytes:]
				default:
					fragment = tr.Message[k*SignatureMessageFragmentSizeInTrytes : (k+1)*SignatureMessageFragmentSizeInTrytes]
				}

				fragments = append(fragments, fragment)
			}
		} else {
			fragments = append(fragments, tr.Message)
		}

		// add output transaction(s) to the bundle for this transfer
		// slice the address in case the user provided one with a checksum
		AddEntry(&bundle, numSignatures, tr.Address[:81], tr.Value, time.Now().Unix(), tr.Tag)

		// sum up the total value to transfer
		total += tr.Value
	}
	return bundle, fragments, total
}

// DoPow computes the nonce field for each transaction so that the last MWM-length trits of the
// transaction hash are all zeroes. Starting from the 0 index transaction, the transactions get chained to
// each other through the trunk transaction hash field. The last transaction in the bundle approves
// the given branch and trunk transactions. This function also initializes the attachment timestamp fields.
func DoPoW(trunkTx, branchTx Trytes, trytes Transactions, mwm uint64, pow pow.PowFunc) error {
	var prev Trytes
	for i := len(trytes) - 1; i >= 0; i-- {
		switch {
		case i == len(trytes)-1:
			trytes[i].TrunkTransaction = trunkTx
			trytes[i].BranchTransaction = branchTx
		default:
			trytes[i].TrunkTransaction = prev
			trytes[i].BranchTransaction = trunkTx
		}

		trytes[i].AttachmentTimestamp = time.Now().UnixNano() / 1000000
		trytes[i].AttachmentTimestampLowerBound = LowerBoundAttachmentTimestamp
		trytes[i].AttachmentTimestampUpperBound = UpperBoundAttachmentTimestamp

		var err error
		trytes[i].Nonce, err = pow(TransactionToTrytes(&trytes[i]), int(mwm))
		if err != nil {
			return err
		}

		prev = trytes[i].Hash
	}
	return nil
}
