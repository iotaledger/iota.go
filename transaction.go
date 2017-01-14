package giota

import (
	"errors"
)

type Transaction struct {
	trytes string
	hash   *Hash

	Hash                     string
	SignatureMessageFragment string
	Address                  string
	Value                    int64 `json:",string"`
	Tag                      string
	Timestamp                int64 `json:",string"`
	CurrentIndex             int64 `json:",string"`
	LastIndex                int64 `json:",string"`
	Bundle                   string
	Nonce                    string
	TrunkTransaction         string
	BranchTransaction        string
}

var (
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	ErrInvalidTransaction     = errors.New("malformed transaction")
)

const (
	SignatureMessageFragmentTrinaryOffset = 0
	SignatureMessageFragmentTrinarySize   = 6561
	AddressTrinaryOffset                  = SignatureMessageFragmentTrinaryOffset + SignatureMessageFragmentTrinarySize
	AddressTrinarySize                    = 243
	ValueTrinaryOffset                    = AddressTrinaryOffset + AddressTrinarySize
	ValueTrinarySize                      = 81
	ValueUsableTrinarySize                = 33
	TagTrinaryOffset                      = ValueTrinaryOffset + ValueTrinarySize
	TagTrinarySize                        = 81
	TimestampTrinaryOffset                = TagTrinaryOffset + TagTrinarySize
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
	NonceTrinaryOffset                    = BranchTransactionTrinaryOffset + BranchTransactionTrinarySize
	NonceTrinarySize                      = 243
)

func TransactionFromBytes(b []int) (Transaction, error) {
	bs := make([]int, 2673)
	copy(bs, b)
	trits := bytesToTrits(bs, 8019)
	return TransactionFromTrits(trits)
}

func TransactionFromTrits(trits []int) (Transaction, error) {
	t := Transaction{}
	t.trytes = tritsToTrytes(trits)

	if t.trytes[2279:2295] != "9999999999999999" {
		return t, errors.New("invalid transaction")
	}

	hash := HashFromTrits(trits)

	t.hash = hash
	t.Hash = t.hash.String()

	t.SignatureMessageFragment = tritsToTrytes(trits[SignatureMessageFragmentTrinaryOffset:SignatureMessageFragmentTrinarySize])
	t.Address = tritsToTrytes(trits[AddressTrinaryOffset : AddressTrinaryOffset+AddressTrinarySize])
	t.Value = tritsToInt(trits[ValueTrinaryOffset : ValueTrinaryOffset+ValueTrinarySize])
	t.Tag = tritsToTrytes(trits[TagTrinaryOffset : TagTrinaryOffset+TagTrinarySize])
	t.Timestamp = tritsToInt(trits[TimestampTrinaryOffset : TimestampTrinaryOffset+TimestampTrinarySize])
	t.CurrentIndex = tritsToInt(trits[CurrentIndexTrinaryOffset : CurrentIndexTrinaryOffset+CurrentIndexTrinarySize])
	t.LastIndex = tritsToInt(trits[LastIndexTrinaryOffset : LastIndexTrinaryOffset+LastIndexTrinarySize])
	t.Bundle = tritsToTrytes(trits[BundleTrinaryOffset : BundleTrinaryOffset+BundleTrinarySize])
	t.TrunkTransaction = tritsToTrytes(trits[TrunkTransactionTrinaryOffset : TrunkTransactionTrinaryOffset+TrunkTransactionTrinarySize])
	t.BranchTransaction = tritsToTrytes(trits[BranchTransactionTrinaryOffset : BranchTransactionTrinaryOffset+BranchTransactionTrinarySize])
	t.Nonce = tritsToTrytes(trits[NonceTrinaryOffset : NonceTrinaryOffset+NonceTrinarySize])

	return t, nil

}

func (t *Transaction) Trytes() string {
	return t.trytes
}

func (t *Transaction) Bytes() []int {
	return tritsToBytes(trytesToTrits(t.Trytes()))
}

func (t *Transaction) Equal(s Transaction) bool {
	return t.Hash == s.Hash &&
		t.SignatureMessageFragment == s.SignatureMessageFragment &&
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
