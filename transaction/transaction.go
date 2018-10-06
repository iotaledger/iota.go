package transaction

import (
	"encoding/json"
	"errors"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"
)

const (
	DefaultMinWeightMagnitude = 14
)

type Transactions []Transaction

// TransactionsToTrytes returns a slice of transaction trytes from the given transactions.
func TransactionsToTrytes(txs Transactions) []Trytes {
	trytes := make([]Trytes, len(txs))
	for i := range txs {
		trytes[i] = TransactionToTrytes(&txs[i])
	}
	return trytes
}

// FinalTransactionTrytes returns a slice of transaction trytes from the given transactions.
// The order of the transactions is reversed in the output slice.
func FinalTransactionTrytes(txs Transactions) []Trytes {
	trytes := TransactionsToTrytes(txs)
	for i, j := 0, len(trytes)-1; i < j; i, j = i+1, j-1 {
		trytes[i], trytes[j] = trytes[j], trytes[i]
	}
	return trytes
}

// Transaction represents a single transaction.
type Transaction struct {
	Hash                          Hash   `json:"hash,string"`
	SignatureMessageFragment      Trytes `json:"currentIndex,string"`
	Address                       Hash   `json:"address,string"`
	Value                         int64  `json:"value,string"`
	ObsoleteTag                   Trytes `json:"obsoleteTag,string"`
	Timestamp                     uint64 `json:"timestamp,string"`
	CurrentIndex                  uint64 `json:"currentIndex,string"`
	LastIndex                     uint64 `json:"lastIndex,string"`
	Bundle                        Hash   `json:"bundle"`
	TrunkTransaction              Hash   `json:"trunkTransaction"`
	BranchTransaction             Hash   `json:"branchTransaction"`
	Tag                           Trytes `json:"tag"`
	AttachmentTimestamp           int64  `json:"attachmentTimestamp,string"`
	AttachmentTimestampLowerBound int64  `json:"attachmentTimestampLowerBound,string"`
	AttachmentTimestampUpperBound int64  `json:"attachmentTimestampUpperBound,string"`
	Nonce                         Trytes `json:"nonce"`
	Confirmed                     *bool  `json:"confirmed,omitempty"`
}

// Trinary sizes and offsets of a transaction
const (
	SignatureMessageFragmentTrinaryOffset = 0
	SignatureMessageFragmentTrinarySize   = 6561
	AddressTrinaryOffset                  = SignatureMessageFragmentTrinaryOffset + SignatureMessageFragmentTrinarySize
	AddressTrinarySize                    = 243
	ValueTrinaryOffset                    = AddressTrinaryOffset + AddressTrinarySize
	ValueTrinarySize                      = 81
	ObsoleteTagTrinaryOffset              = ValueTrinaryOffset + ValueTrinarySize
	ObsoleteTagTrinarySize                = 81
	TimestampTrinaryOffset                = ObsoleteTagTrinaryOffset + ObsoleteTagTrinarySize
	TimestampTrinarySize                  = 27
	CurrentIndexTrinaryOffset             = TimestampTrinaryOffset + TimestampTrinarySize
	CurrentIndexTrinarySize               = 27
	LastIndexTrinaryOffset                = CurrentIndexTrinaryOffset + CurrentIndexTrinarySize
	LastIndexTrinarySize                  = 27
	BundleTrinaryOffset                   = LastIndexTrinaryOffset + LastIndexTrinarySize
	BundleTrinarySize                     = 243
	TrunkTransactionTrinaryOffset         = BundleTrinaryOffset + BundleTrinarySize
	TrunkTransactionTrinarySize           = 243
	BranchTransactionTrinaryOffset        = TrunkTransactionTrinaryOffset + TrunkTransactionTrinarySize
	BranchTransactionTrinarySize          = 243
	TagTrinaryOffset                      = BranchTransactionTrinaryOffset + BranchTransactionTrinarySize
	TagTrinarySize                        = 81
	AttachmentTimestampTrinaryOffset      = TagTrinaryOffset + TagTrinarySize
	AttachmentTimestampTrinarySize        = 27

	AttachmentTimestampLowerBoundTrinaryOffset = AttachmentTimestampTrinaryOffset + AttachmentTimestampTrinarySize
	AttachmentTimestampLowerBoundTrinarySize   = 27
	AttachmentTimestampUpperBoundTrinaryOffset = AttachmentTimestampLowerBoundTrinaryOffset + AttachmentTimestampLowerBoundTrinarySize
	AttachmentTimestampUpperBoundTrinarySize   = 27
	NonceTrinaryOffset                         = AttachmentTimestampUpperBoundTrinaryOffset + AttachmentTimestampUpperBoundTrinarySize
	NonceTrinarySize                           = 81

	TransactionTrinarySize = SignatureMessageFragmentTrinarySize + AddressTrinarySize +
		ValueTrinarySize + ObsoleteTagTrinarySize + TimestampTrinarySize +
		CurrentIndexTrinarySize + LastIndexTrinarySize + BundleTrinarySize +
		TrunkTransactionTrinarySize + BranchTransactionTrinarySize +
		TagTrinarySize + AttachmentTimestampTrinarySize +
		AttachmentTimestampLowerBoundTrinarySize + AttachmentTimestampUpperBoundTrinarySize +
		NonceTrinarySize
)

// NewTransaction makes a new transaction from the given trytes.
func NewTransaction(trytes Trytes) (*Transaction, error) {
	var t *Transaction
	var err error
	if err := ValidTransaction(trytes); err != nil {
		return nil, err
	}

	if t, err = ParseTransaction(TrytesToTrits(trytes)); err != nil {
		return nil, err
	}

	return t, nil
}

// AsTransactionObjects constructs new transactions from the given raw trytes.
func AsTransactionObjects(rawTrytes []Trytes, hashes Hashes) (Transactions, error) {
	txs := Transactions{}
	for i := range rawTrytes {
		tx, err := NewTransaction(rawTrytes[i])
		if err != nil {
			return nil, err
		}
		if hashes != nil {
			tx.Hash = hashes[i]
		}
		txs = append(txs, *tx)
	}
	return txs, nil
}

// ValidTransaction checks whether the given trytes make up a valid transaction.
func ValidTransaction(trytes Trytes) error {
	err := ValidTrytes(trytes)

	switch {
	case err != nil:
		return errors.New("invalid transaction " + err.Error())
	case len(trytes) != TransactionTrinarySize/3:
		return errors.New("invalid trits counts in transaction")
	case trytes[2279:2295] != "9999999999999999":
		return errors.New("invalid value in transaction")
	default:
		return nil
	}
}

func ParseTransaction(trits Trits) (*Transaction, error) {
	var err error
	t := &Transaction{}
	t.SignatureMessageFragment = MustTritsToTrytes(trits[SignatureMessageFragmentTrinaryOffset:SignatureMessageFragmentTrinarySize])
	t.Address, err = TritsToTrytes(trits[AddressTrinaryOffset : AddressTrinaryOffset+AddressTrinarySize])
	if err != nil {
		return nil, err
	}
	t.Value = TritsToInt(trits[ValueTrinaryOffset : ValueTrinaryOffset+ValueTrinarySize])
	t.ObsoleteTag = MustTritsToTrytes(trits[ObsoleteTagTrinaryOffset : ObsoleteTagTrinaryOffset+ObsoleteTagTrinarySize])
	t.Timestamp = uint64(TritsToInt(trits[TimestampTrinaryOffset : TimestampTrinaryOffset+TimestampTrinarySize]))
	t.CurrentIndex = uint64(TritsToInt(trits[CurrentIndexTrinaryOffset : CurrentIndexTrinaryOffset+CurrentIndexTrinarySize]))
	t.LastIndex = uint64(TritsToInt(trits[LastIndexTrinaryOffset : LastIndexTrinaryOffset+LastIndexTrinarySize]))
	t.Bundle = MustTritsToTrytes(trits[BundleTrinaryOffset : BundleTrinaryOffset+BundleTrinarySize])
	t.TrunkTransaction = MustTritsToTrytes(trits[TrunkTransactionTrinaryOffset : TrunkTransactionTrinaryOffset+TrunkTransactionTrinarySize])
	t.BranchTransaction = MustTritsToTrytes(trits[BranchTransactionTrinaryOffset : BranchTransactionTrinaryOffset+BranchTransactionTrinarySize])
	t.Tag = MustTritsToTrytes(trits[TagTrinaryOffset : TagTrinaryOffset+TagTrinarySize])
	t.AttachmentTimestamp = TritsToInt(trits[AttachmentTimestampTrinaryOffset : AttachmentTimestampTrinaryOffset+AttachmentTimestampTrinarySize])
	t.AttachmentTimestampLowerBound = TritsToInt(trits[AttachmentTimestampLowerBoundTrinaryOffset : AttachmentTimestampLowerBoundTrinaryOffset+AttachmentTimestampLowerBoundTrinarySize])
	t.AttachmentTimestampUpperBound = TritsToInt(trits[AttachmentTimestampUpperBoundTrinaryOffset : AttachmentTimestampUpperBoundTrinaryOffset+AttachmentTimestampUpperBoundTrinarySize])
	t.Nonce = MustTritsToTrytes(trits[NonceTrinaryOffset : NonceTrinaryOffset+NonceTrinarySize])
	return t, nil
}

// Trytes converts the transaction to Trytes.
func TransactionToTrytes(t *Transaction) Trytes {
	tr := make(Trits, TransactionTrinarySize)
	copy(tr, TrytesToTrits(t.SignatureMessageFragment))
	copy(tr[AddressTrinaryOffset:], TrytesToTrits(t.Address))
	copy(tr[ValueTrinaryOffset:], IntToTrits(t.Value))
	copy(tr[ObsoleteTagTrinaryOffset:], TrytesToTrits(t.ObsoleteTag))
	copy(tr[TimestampTrinaryOffset:], IntToTrits(int64(t.Timestamp)))
	copy(tr[CurrentIndexTrinaryOffset:], IntToTrits(int64(t.CurrentIndex)))
	copy(tr[LastIndexTrinaryOffset:], IntToTrits(int64(t.LastIndex)))
	copy(tr[BundleTrinaryOffset:], TrytesToTrits(t.Bundle))
	copy(tr[TrunkTransactionTrinaryOffset:], TrytesToTrits(t.TrunkTransaction))
	copy(tr[BranchTransactionTrinaryOffset:], TrytesToTrits(t.BranchTransaction))
	copy(tr[TagTrinaryOffset:], TrytesToTrits(t.Tag))
	copy(tr[AttachmentTimestampTrinaryOffset:], IntToTrits(t.AttachmentTimestamp))
	copy(tr[AttachmentTimestampLowerBoundTrinaryOffset:], IntToTrits(t.AttachmentTimestampLowerBound))
	copy(tr[AttachmentTimestampUpperBoundTrinaryOffset:], IntToTrits(t.AttachmentTimestampUpperBound))
	copy(tr[NonceTrinaryOffset:], TrytesToTrits(t.Nonce))
	return MustTritsToTrytes(tr)
}

// TransactionHash makes a transaction hash from the given transaction.
func TransactionHash(t *Transaction) Hash {
	return curl.Hash(TransactionToTrytes(t))
}

// HasValidNonce checks if the transaction has the valid MinWeightMagnitude.
// In order to check the MWM we count trailing 0's of the curlp hash of a transaction.
func HasValidNonce(t *Transaction, mwm uint64) bool {
	return TrailingZeros(TrytesToTrits(TransactionHash(t))) >= int64(mwm)
}

// UnmarshalJSON makes transaction struct from json.
func UnmarshalJSON(b []byte) (*Transaction, error) {
	var s Trytes
	var err error
	if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	if err = ValidTransaction(s); err != nil {
		return nil, err
	}

	return ParseTransaction(TrytesToTrits(s))
}

// MarshalJSON makes trytes ([]byte) from a transaction.
func MarshalJSON(t *Transaction) ([]byte, error) {
	return json.Marshal(t)
}

// IsTailTransaction checks if given transaction object is tail transaction.
// A tail transaction is one with currentIndex = 0.
func IsTailTransaction(t *Transaction) bool {
	return t.CurrentIndex == 0
}
