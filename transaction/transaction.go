package transaction

import (
	"encoding/json"
	"errors"
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/signing"
	"github.com/iotaledger/giota/trinary"
	"time"
)

const (
	EmptyTag                  = trinary.Trytes("999999999999999999999999999")
	DefaultMinWeightMagnitude = 14
)

type Transactions []Transaction

// Transaction contains all info needed for an iota transaction
type Transaction struct {
	SignatureMessageFragment      trinary.Trytes
	Address                       signing.Address
	Value                         int64 `json:",string"`
	ObsoleteTag                   trinary.Trytes
	Timestamp                     time.Time `json:",string"`
	CurrentIndex                  int64     `json:",string"`
	LastIndex                     int64     `json:",string"`
	Bundle                        trinary.Trytes
	TrunkTransaction              trinary.Trytes
	BranchTransaction             trinary.Trytes
	Tag                           trinary.Trytes
	AttachmentTimestamp           trinary.Trytes
	AttachmentTimestampLowerBound trinary.Trytes
	AttachmentTimestampUpperBound trinary.Trytes
	Nonce                         trinary.Trytes
}

// errors for tx
var (
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	ErrInvalidTransaction     = errors.New("malformed transaction")
)

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
func NewTransaction(trytes trinary.Trytes) (*Transaction, error) {
	t := Transaction{}
	if err := checkTransaction(trytes); err != nil {
		return nil, err
	}

	if err := t.parser(trytes.Trits()); err != nil {
		return nil, err
	}

	return &t, nil
}

// AsTransactionObjects constructs new transactions from the given raw trytes.
func AsTransactionObjects(rawTrytes ...trinary.Trytes) (Transactions, error) {
	txs := Transactions{}
	for i := range rawTrytes {
		tx, err := NewTransaction(rawTrytes[i])
		if err != nil {
			return nil, err
		}
		txs = append(txs, *tx)
	}
	return txs, nil
}

func checkTransaction(trytes trinary.Trytes) error {
	err := trytes.IsValid()

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

func (t *Transaction) parser(trits trinary.Trits) error {
	var err error
	t.SignatureMessageFragment = trits[SignatureMessageFragmentTrinaryOffset:SignatureMessageFragmentTrinarySize].MustTrytes()
	t.Address, err = signing.ToAddress(trits[AddressTrinaryOffset : AddressTrinaryOffset+AddressTrinarySize].MustTrytes())
	if err != nil {
		return err
	}
	t.Value = trits[ValueTrinaryOffset : ValueTrinaryOffset+ValueTrinarySize].Int()
	t.ObsoleteTag = trits[ObsoleteTagTrinaryOffset : ObsoleteTagTrinaryOffset+ObsoleteTagTrinarySize].MustTrytes()
	timestamp := trits[TimestampTrinaryOffset : TimestampTrinaryOffset+TimestampTrinarySize].Int()
	t.Timestamp = time.Unix(timestamp, 0)
	t.CurrentIndex = trits[CurrentIndexTrinaryOffset : CurrentIndexTrinaryOffset+CurrentIndexTrinarySize].Int()
	t.LastIndex = trits[LastIndexTrinaryOffset : LastIndexTrinaryOffset+LastIndexTrinarySize].Int()
	t.Bundle = trits[BundleTrinaryOffset : BundleTrinaryOffset+BundleTrinarySize].MustTrytes()
	t.TrunkTransaction = trits[TrunkTransactionTrinaryOffset : TrunkTransactionTrinaryOffset+TrunkTransactionTrinarySize].MustTrytes()
	t.BranchTransaction = trits[BranchTransactionTrinaryOffset : BranchTransactionTrinaryOffset+BranchTransactionTrinarySize].MustTrytes()
	t.Tag = trits[TagTrinaryOffset : TagTrinaryOffset+TagTrinarySize].MustTrytes()
	t.AttachmentTimestamp = trits[AttachmentTimestampTrinaryOffset : AttachmentTimestampTrinaryOffset+AttachmentTimestampTrinarySize].MustTrytes()
	t.AttachmentTimestampLowerBound = trits[AttachmentTimestampLowerBoundTrinaryOffset : AttachmentTimestampLowerBoundTrinaryOffset+AttachmentTimestampLowerBoundTrinarySize].MustTrytes()
	t.AttachmentTimestampUpperBound = trits[AttachmentTimestampUpperBoundTrinaryOffset : AttachmentTimestampUpperBoundTrinaryOffset+AttachmentTimestampUpperBoundTrinarySize].MustTrytes()
	t.Nonce = trits[NonceTrinaryOffset : NonceTrinaryOffset+NonceTrinarySize].MustTrytes()

	return nil
}

// Trytes converts the transaction to Trytes.
func (t *Transaction) Trytes() trinary.Trytes {
	tr := make(trinary.Trits, TransactionTrinarySize)
	copy(tr, t.SignatureMessageFragment.Trits())
	copy(tr[AddressTrinaryOffset:], trinary.Trytes(t.Address).Trits())
	copy(tr[ValueTrinaryOffset:], trinary.IntToTrits(t.Value, ValueTrinarySize))
	copy(tr[ObsoleteTagTrinaryOffset:], t.ObsoleteTag.Trits())
	copy(tr[TimestampTrinaryOffset:], trinary.IntToTrits(t.Timestamp.Unix(), TimestampTrinarySize))
	copy(tr[CurrentIndexTrinaryOffset:], trinary.IntToTrits(t.CurrentIndex, CurrentIndexTrinarySize))
	copy(tr[LastIndexTrinaryOffset:], trinary.IntToTrits(t.LastIndex, LastIndexTrinarySize))
	copy(tr[BundleTrinaryOffset:], t.Bundle.Trits())
	copy(tr[TrunkTransactionTrinaryOffset:], t.TrunkTransaction.Trits())
	copy(tr[BranchTransactionTrinaryOffset:], t.BranchTransaction.Trits())
	copy(tr[TagTrinaryOffset:], t.Tag.Trits())
	copy(tr[AttachmentTimestampTrinaryOffset:], t.AttachmentTimestamp.Trits())
	copy(tr[AttachmentTimestampLowerBoundTrinaryOffset:], t.AttachmentTimestampLowerBound.Trits())
	copy(tr[AttachmentTimestampUpperBoundTrinaryOffset:], t.AttachmentTimestampUpperBound.Trits())
	copy(tr[NonceTrinaryOffset:], t.Nonce.Trits())
	return tr.MustTrytes()
}

// HasValidNonce checks if the transaction has the valid MinWeightMagnitude.
// In order to check the MWM we count trailing 0's of the curlp hash of a
// transaction.
func (t *Transaction) HasValidNonce(mwm int64) bool {
	return t.Hash().Trits().TrailingZeros() >= mwm
}

// Hash returns the hash of the transaction.
func (t *Transaction) Hash() trinary.Trytes {
	return curl.Hash(t.Trytes())
}

// UnmarshalJSON makes transaction struct from json.
func (t *Transaction) UnmarshalJSON(b []byte) error {
	var s trinary.Trytes
	var err error

	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	if err = checkTransaction(s); err != nil {
		return err
	}

	return t.parser(s.Trits())
}

// MarshalJSON makes trytes ([]byte) from a transaction.
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Trytes() + `"`), nil
}

// Checks if given transaction object is tail transaction.
// A tail transaction is one with currentIndex=0.
func (t *Transaction) IsTail() bool {
	return t.CurrentIndex == 0
}