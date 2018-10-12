package transaction

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/utils"
	"github.com/pkg/errors"
)

type Transactions []Transaction

// Transaction represents a single transaction.
type Transaction struct {
	Hash                          Hash   `json:"hash"`
	SignatureMessageFragment      Trytes `json:"signatureMessageFragment"`
	Address                       Hash   `json:"address"`
	Value                         int64  `json:"value"`
	ObsoleteTag                   Trytes `json:"obsoleteTag"`
	Timestamp                     uint64 `json:"timestamp"`
	CurrentIndex                  uint64 `json:"currentIndex"`
	LastIndex                     uint64 `json:"lastIndex"`
	Bundle                        Hash   `json:"bundle"`
	TrunkTransaction              Hash   `json:"trunkTransaction"`
	BranchTransaction             Hash   `json:"branchTransaction"`
	Tag                           Trytes `json:"tag"`
	AttachmentTimestamp           int64  `json:"attachmentTimestamp"`
	AttachmentTimestampLowerBound int64  `json:"attachmentTimestampLowerBound"`
	AttachmentTimestampUpperBound int64  `json:"attachmentTimestampUpperBound"`
	Nonce                         Trytes `json:"nonce"`
	Confirmed                     *bool  `json:"confirmed,omitempty"`
}

// ParseTransaction parses the trits and returns a transaction object.
// The trits slice must be TransactionTrinarySize in length.
func ParseTransaction(trits Trits) (*Transaction, error) {
	var err error

	if len(trits) != TransactionTrinarySize {
		return nil, ErrInvalidTransaction
	}

	if err := ValidTrits(trits); err != nil {
		return nil, err
	}

	t := &Transaction{}
	t.SignatureMessageFragment = MustTritsToTrytes(trits[SignatureMessageFragmentTrinaryOffset:SignatureMessageFragmentTrinarySize])
	t.Address, err = TritsToTrytes(trits[AddressTrinaryOffset : AddressTrinaryOffset+AddressTrinarySize])
	if err != nil {
		return nil, err
	}
	t.Value = TritsToInt(trits[ValueOffsetTrinary : ValueOffsetTrinary+ValueSizeTrinary])
	t.ObsoleteTag = MustTritsToTrytes(trits[ObsoleteTagTrinaryOffset : ObsoleteTagTrinaryOffset+ObsoleteTagTrinarySize])
	t.Timestamp = uint64(TritsToInt(trits[TimestampTrinaryOffset : TimestampTrinaryOffset+TimestampTrinarySize]))
	t.CurrentIndex = uint64(TritsToInt(trits[CurrentIndexTrinaryOffset : CurrentIndexTrinaryOffset+CurrentIndexTrinarySize]))
	t.LastIndex = uint64(TritsToInt(trits[LastIndexTrinaryOffset : LastIndexTrinaryOffset+LastIndexTrinarySize]))
	if t.CurrentIndex > t.LastIndex {
		return nil, errors.Wrap(ErrInvalidIndex, "current index is bigger than last index")
	}
	t.Bundle = MustTritsToTrytes(trits[BundleTrinaryOffset : BundleTrinaryOffset+BundleTrinarySize])
	t.TrunkTransaction = MustTritsToTrytes(trits[TrunkTransactionTrinaryOffset : TrunkTransactionTrinaryOffset+TrunkTransactionTrinarySize])
	t.BranchTransaction = MustTritsToTrytes(trits[BranchTransactionTrinaryOffset : BranchTransactionTrinaryOffset+BranchTransactionTrinarySize])
	t.Tag = MustTritsToTrytes(trits[TagTrinaryOffset : TagTrinaryOffset+TagTrinarySize])
	t.AttachmentTimestamp = TritsToInt(trits[AttachmentTimestampTrinaryOffset : AttachmentTimestampTrinaryOffset+AttachmentTimestampTrinarySize])
	t.AttachmentTimestampLowerBound = TritsToInt(trits[AttachmentTimestampLowerBoundTrinaryOffset : AttachmentTimestampLowerBoundTrinaryOffset+AttachmentTimestampLowerBoundTrinarySize])
	t.AttachmentTimestampUpperBound = TritsToInt(trits[AttachmentTimestampUpperBoundTrinaryOffset : AttachmentTimestampUpperBoundTrinaryOffset+AttachmentTimestampUpperBoundTrinarySize])
	t.Nonce = MustTritsToTrytes(trits[NonceTrinaryOffset : NonceTrinaryOffset+NonceTrinarySize])
	t.Hash = TransactionHash(t)

	return t, nil
}

// ValidTransactionTrytes checks whether the given trytes make up a valid transaction schematically.
func ValidTransactionTrytes(trytes Trytes) error {
	// verifies length and trytes values
	if !utils.IsTrytesOfExactLength(trytes, TransactionTrinarySize/3) {
		return ErrInvalidTrytes
	}

	if trytes[2279:2295] != "9999999999999999" {
		return ErrInvalidTrytes
	}

	return nil
}

// NewTransaction makes a new transaction from the given trytes.
func NewTransaction(trytes Trytes) (*Transaction, error) {
	var t *Transaction
	var err error

	if err := ValidTransactionTrytes(trytes); err != nil {
		return nil, err
	}

	if t, err = ParseTransaction(MustTrytesToTrits(trytes)); err != nil {
		return nil, err
	}

	return t, nil
}

// TransactionToTrytes converts the transaction to trytes.
func TransactionToTrytes(t *Transaction) (Trytes, error) {
	tr := make(Trits, TransactionTrinarySize)
	if !utils.IsTrytesOfExactLength(t.SignatureMessageFragment, SignatureMessageFragmentTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid signature message fragment")
	}
	copy(tr, MustTrytesToTrits(t.SignatureMessageFragment))

	if !utils.IsTrytesOfExactLength(t.Address, AddressTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid address")
	}
	copy(tr[AddressTrinaryOffset:], MustTrytesToTrits(t.Address))

	copy(tr[ValueOffsetTrinary:], IntToTrits(t.Value))
	if !utils.IsTrytesOfExactLength(t.ObsoleteTag, ObsoleteTagTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid obsolete tag")
	}
	copy(tr[ObsoleteTagTrinaryOffset:], MustTrytesToTrits(t.ObsoleteTag))

	copy(tr[TimestampTrinaryOffset:], IntToTrits(int64(t.Timestamp)))
	if t.CurrentIndex > t.LastIndex {
		return "", errors.Wrap(ErrInvalidIndex, "current index is bigger than last index")
	}

	copy(tr[CurrentIndexTrinaryOffset:], IntToTrits(int64(t.CurrentIndex)))
	copy(tr[LastIndexTrinaryOffset:], IntToTrits(int64(t.LastIndex)))
	if !utils.IsTrytesOfExactLength(t.Bundle, BundleTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid bundle hash")
	}
	copy(tr[BundleTrinaryOffset:], MustTrytesToTrits(t.Bundle))

	if !utils.IsTrytesOfExactLength(t.TrunkTransaction, TrunkTransactionTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid trunk tx hash")
	}
	copy(tr[TrunkTransactionTrinaryOffset:], MustTrytesToTrits(t.TrunkTransaction))

	if !utils.IsTrytesOfExactLength(t.BranchTransaction, BranchTransactionTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid branch tx hash")
	}
	copy(tr[BranchTransactionTrinaryOffset:], MustTrytesToTrits(t.BranchTransaction))

	if !utils.IsTrytesOfExactLength(t.Tag, TagTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid tag")
	}
	copy(tr[TagTrinaryOffset:], MustTrytesToTrits(t.Tag))
	copy(tr[AttachmentTimestampTrinaryOffset:], IntToTrits(t.AttachmentTimestamp))
	copy(tr[AttachmentTimestampLowerBoundTrinaryOffset:], IntToTrits(t.AttachmentTimestampLowerBound))
	copy(tr[AttachmentTimestampUpperBoundTrinaryOffset:], IntToTrits(t.AttachmentTimestampUpperBound))
	if !utils.IsTrytesOfExactLength(t.Nonce, NonceTrinarySize/3) {
		return "", errors.Wrap(ErrInvalidTrytes, "invalid nonce")
	}
	copy(tr[NonceTrinaryOffset:], MustTrytesToTrits(t.Nonce))

	return MustTritsToTrytes(tr), nil
}

// MustTransactionToTrytes converts the transaction to trytes.
func MustTransactionToTrytes(t *Transaction) Trytes {
	trytes, err := TransactionToTrytes(t)
	if err != nil {
		panic(err)
	}
	return trytes
}

// TransactionsToTrytes returns a slice of transaction trytes from the given transactions.
func TransactionsToTrytes(txs Transactions) ([]Trytes, error) {
	trytes := make([]Trytes, len(txs))
	var err error
	for i := range txs {
		trytes[i], err = TransactionToTrytes(&txs[i])
		if err != nil {
			return nil, err
		}
	}
	return trytes, nil
}

// MustTransactionsToTrytes returns a slice of transaction trytes from the given transactions.
func MustTransactionsToTrytes(txs Transactions) []Trytes {
	trytes := make([]Trytes, len(txs))
	for i := range txs {
		trytes[i] = MustTransactionToTrytes(&txs[i])
	}
	return trytes
}

// FinalTransactionTrytes returns a slice of transaction trytes from the given transactions.
// The order of the transactions is reversed in the output slice.
func FinalTransactionTrytes(txs Transactions) ([]Trytes, error) {
	trytes, err := TransactionsToTrytes(txs)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(trytes)-1; i < j; i, j = i+1, j-1 {
		trytes[i], trytes[j] = trytes[j], trytes[i]
	}
	return trytes, nil
}

// MustFinalTransactionTrytes returns a slice of transaction trytes from the given transactions.
// The order of the transactions is reversed in the output slice.
func MustFinalTransactionTrytes(txs Transactions) []Trytes {
	trytes := MustTransactionsToTrytes(txs)
	for i, j := 0, len(trytes)-1; i < j; i, j = i+1, j-1 {
		trytes[i], trytes[j] = trytes[j], trytes[i]
	}
	return trytes
}

// AsTransactionObjects constructs new transactions from the given raw trytes.
func AsTransactionObjects(rawTrytes []Trytes, hashes Hashes) (Transactions, error) {
	txs := Transactions{}
	for i := range rawTrytes {
		tx, err := NewTransaction(rawTrytes[i])
		if err != nil {
			return nil, err
		}
		if hashes != nil && len(hashes) > 0 {
			tx.Hash = hashes[i]
		}
		txs = append(txs, *tx)
	}
	return txs, nil
}

// TransactionHash makes a transaction hash from the given transaction.
func TransactionHash(t *Transaction) Hash {
	return curl.HashTrytes(MustTransactionToTrytes(t))
}

// HasValidNonce checks if the transaction has the valid MinWeightMagnitude.
// MWM corresponds to the amount of zero trits at the end of the transaction hash.
func HasValidNonce(t *Transaction, mwm uint64) bool {
	return TrailingZeros(MustTrytesToTrits(TransactionHash(t))) >= int64(mwm)
}

// IsTailTransaction checks if given transaction object is tail transaction.
// A tail transaction is one with currentIndex = 0.
func IsTailTransaction(t *Transaction) bool {
	return t.CurrentIndex == 0
}
