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

	t.SignatureMessageFragment = tritsToTrytes(trits[0:6561])
	t.Value = tritsToInt(trits[6804 : 6804+33])
	t.Tag = tritsToTrytes(trits[6885 : 6885+81])
	t.Timestamp = tritsToInt(trits[6966 : 6966+27])
	t.CurrentIndex = tritsToInt(trits[6993 : 6993+27])
	t.LastIndex = tritsToInt(trits[7020 : 7020+27])
	t.Bundle = tritsToTrytes(trits[7047 : 7047+243])
	t.TrunkTransaction = tritsToTrytes(trits[7290 : 7290+243])
	t.BranchTransaction = tritsToTrytes(trits[7533 : 7533+243])
	t.Nonce = tritsToTrytes(trits[7776 : 7776+243])

	t.Address = tritsToTrytes(trits[6561 : 6561+243])

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
