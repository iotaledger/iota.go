/*
MIT License

Copyright (c) 2016 Sascha Hanse
Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package giota

import (
	"encoding/json"
	"errors"
	"time"
)

//Transaction is transaction  structure for iota.
type Transaction struct {
	SignatureMessageFragment Trytes
	Address                  Address
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
)

//NewTransaction makes tx from trits.
func NewTransaction(trytes Trytes) (*Transaction, error) {
	t := Transaction{}
	if err := checkTx(trytes); err != nil {
		return nil, err
	}
	err := t.parser(trytes.Trits())
	return &t, err
}

func checkTx(trytes Trytes) error {
	if err := trytes.IsValid(); err != nil {
		return errors.New("invalid transaction " + err.Error())
	}
	if len(trytes) != transactionTrinarySize/3 {
		return errors.New("invalid trits counts in transaction")
	}

	if trytes[2279:2295] != "9999999999999999" {
		return errors.New("invalid value in transaction")
	}
	return nil
}

func (t *Transaction) parser(trits Trits) error {
	var err error
	t.SignatureMessageFragment = trits[signatureMessageFragmentTrinaryOffset:signatureMessageFragmentTrinarySize].Trytes()
	t.Address, err = trits[addressTrinaryOffset : addressTrinaryOffset+addressTrinarySize].Trytes().ToAddress()
	if err != nil {
		return err
	}
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

	return nil
}

//Trytes converts the transaction to Trytes.
func (t *Transaction) Trytes() Trytes {
	tr := make(Trits, transactionTrinarySize)
	copy(tr, t.SignatureMessageFragment.Trits())
	copy(tr[addressTrinaryOffset:], Trytes(t.Address).Trits())
	copy(tr[valueTrinaryOffset:], Int2Trits(t.Value, valueTrinarySize))
	copy(tr[tagTrinaryOffset:], t.Tag.Trits())
	copy(tr[timestampTrinaryOffset:], Int2Trits(t.Timestamp.Unix(), timestampTrinarySize))
	copy(tr[currentIndexTrinaryOffset:], Int2Trits(t.CurrentIndex, currentIndexTrinarySize))
	copy(tr[lastIndexTrinaryOffset:], Int2Trits(t.LastIndex, lastIndexTrinarySize))
	copy(tr[bundleTrinaryOffset:], t.Bundle.Trits())
	copy(tr[trunkTransactionTrinaryOffset:], t.TrunkTransaction.Trits())
	copy(tr[branchTransactionTrinaryOffset:], t.BranchTransaction.Trits())
	copy(tr[nonceTrinaryOffset:], t.Nonce.Trits())
	return tr.Trytes()
}

//HasValidNonce checks t's hash has valid MinWeightMagnitude.
func (t *Transaction) HasValidNonce() bool {
	h := t.Hash()
	for i := len(h) - 1; i > len(h)-1-MinWeightMagnitude/3; i-- {
		if h[i] != '9' {
			return false
		}
	}
	return true
}

//Hash returns the hash of the transaction.
func (t *Transaction) Hash() Trytes {
	return t.Trytes().Hash()
}

//UnmarshalJSON makes transaction struct from json.
func (t *Transaction) UnmarshalJSON(b []byte) error {
	var s Trytes
	var err error
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}
	if err = t.parser(s.Trits()); err != nil {
		return err
	}
	return nil
}

//MarshalJSON makes trytes ([]byte) from a transaction.
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Trytes() + `"`), nil
}
