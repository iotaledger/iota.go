package giota

import (
	"errors"
	"time"
)

//Transaction is transaction  structure for iota.
type Transaction struct {
	SignatureMessageFragment Trytes
	Address                  Trytes
	Value                    int64 `json:",string"`
	Tag                      Trytes
	Timestamp                time.Time `json:",string"`
	CurrentIndex             int64     `json:",string"`
	LastIndex                int64     `json:",string"`
	Bundle                   Trytes
	TrunkTransaction         Trytes
	BranchTransaction        Trytes
	Nonce                    Trytes
}

//errors for tx.
var (
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	ErrInvalidTransaction     = errors.New("malformed transaction")
)

//sizes and offsets of tx.
const (
	signatureMessageFragmentTrinaryOffset = 0
	signatureMessageFragmentTrinarySize   = 6561
	addressTrinaryOffset                  = signatureMessageFragmentTrinaryOffset + signatureMessageFragmentTrinarySize
	addressTrinarySize                    = 243
	valueTrinaryOffset                    = addressTrinaryOffset + addressTrinarySize
	valueTrinarySize                      = 81
	valueUsableTrinarySize                = 33
	tagTrinaryOffset                      = valueTrinaryOffset + valueTrinarySize
	tagTrinarySize                        = 81
	timestampTrinaryOffset                = tagTrinaryOffset + tagTrinarySize
	timestampTrinarySize                  = 27
	currentIndexTrinaryOffset             = timestampTrinaryOffset + timestampTrinarySize
	currentIndexTrinarySize               = 27
	lastIndexTrinaryOffset                = currentIndexTrinaryOffset + currentIndexTrinarySize
	lastIndexTrinarySize                  = 27
	bundleTrinaryOffset                   = lastIndexTrinaryOffset + lastIndexTrinarySize
	bundleTrinarySize                     = 243
	trunkTransactionTrinaryOffset         = bundleTrinaryOffset + bundleTrinarySize
	trunkTransactionTrinarySize           = 243
	branchTransactionTrinaryOffset        = trunkTransactionTrinaryOffset + trunkTransactionTrinarySize
	branchTransactionTrinarySize          = 243
	nonceTrinaryOffset                    = branchTransactionTrinaryOffset + branchTransactionTrinarySize
	nonceTrinarySize                      = 243

	transactionTrinarySize = signatureMessageFragmentTrinarySize + addressTrinarySize +
		valueTrinarySize + tagTrinarySize + timestampTrinarySize +
		currentIndexTrinarySize + lastIndexTrinarySize + bundleTrinarySize +
		trunkTransactionTrinarySize + branchTransactionTrinarySize +
		nonceTrinarySize

	transactionTryteSize = transactionTrinarySize / NumberOfTritsPerTryte
)

//NewTransaction makes tx from trits.
func NewTransaction(trits Trits) (*Transaction, error) {
	t := Transaction{}
	trytes := trits.Trytes()

	if err := trits.IsValid(); err != nil || len(trits) != transactionTrinarySize {
		return nil, errors.New("invalid transaction " + err.Error())
	}

	if trytes[2279:2295] != "9999999999999999" {
		return nil, errors.New("invalid transaction")
	}

	t.SignatureMessageFragment = trits[signatureMessageFragmentTrinaryOffset:signatureMessageFragmentTrinarySize].Trytes()
	t.Address = trits[addressTrinaryOffset : addressTrinaryOffset+addressTrinarySize].Trytes()
	t.Value = trits[valueTrinaryOffset : valueTrinaryOffset+valueTrinarySize].Int()
	t.Tag = trits[tagTrinaryOffset : tagTrinaryOffset+tagTrinarySize].Trytes()
	timestamp := trits[timestampTrinaryOffset : timestampTrinaryOffset+timestampTrinarySize].Int()
	t.Timestamp = time.Unix(timestamp, 0)
	t.CurrentIndex = trits[currentIndexTrinaryOffset : currentIndexTrinaryOffset+currentIndexTrinarySize].Int()
	t.LastIndex = trits[lastIndexTrinaryOffset : lastIndexTrinaryOffset+lastIndexTrinarySize].Int()
	t.Bundle = trits[bundleTrinaryOffset : bundleTrinaryOffset+bundleTrinarySize].Trytes()
	t.TrunkTransaction = trits[trunkTransactionTrinaryOffset : trunkTransactionTrinaryOffset+trunkTransactionTrinarySize].Trytes()
	t.BranchTransaction = trits[branchTransactionTrinaryOffset : branchTransactionTrinaryOffset+branchTransactionTrinarySize].Trytes()
	t.Nonce = trits[nonceTrinaryOffset : nonceTrinaryOffset+nonceTrinarySize].Trytes()

	return &t, nil

}

//Trits returns trits representation of t.
func (t *Transaction) Trits() Trits {
	tr := make(Trits, transactionTrinarySize)
	copy(tr, t.SignatureMessageFragment.Trits())
	copy(tr[addressTrinaryOffset:], t.Address.Trits())
	copy(tr[valueTrinaryOffset:], Int2Trits(t.Value, valueTrinarySize))
	copy(tr[tagTrinaryOffset:], t.Tag.Trits())
	copy(tr[timestampTrinaryOffset:], Int2Trits(t.Timestamp.Unix(), timestampTrinarySize))
	copy(tr[currentIndexTrinaryOffset:], Int2Trits(t.CurrentIndex, currentIndexTrinarySize))
	copy(tr[lastIndexTrinaryOffset:], Int2Trits(t.LastIndex, lastIndexTrinarySize))
	copy(tr[bundleTrinaryOffset:], t.Bundle.Trits())
	copy(tr[trunkTransactionTrinaryOffset:], t.TrunkTransaction.Trits())
	copy(tr[branchTransactionTrinaryOffset:], t.BranchTransaction.Trits())
	copy(tr[nonceTrinaryOffset:], t.Nonce.Trits())
	return tr
}

//Equal returns true if t==s
func (t *Transaction) Equal(s *Transaction) bool {
	return t.SignatureMessageFragment == s.SignatureMessageFragment &&
		t.Address == s.Address &&
		t.Value == s.Value &&
		t.Tag == s.Tag &&
		t.Timestamp == s.Timestamp &&
		t.CurrentIndex == s.CurrentIndex &&
		t.LastIndex == s.LastIndex &&
		t.Bundle == s.Bundle &&
		t.Nonce == s.Nonce &&
		t.TrunkTransaction == s.TrunkTransaction &&
		t.BranchTransaction == s.BranchTransaction
}
