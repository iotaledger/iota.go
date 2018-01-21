package mamgoiota

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/mamgoiota/mamutils"

	"github.com/iotaledger/giota"
	"github.com/stretchr/testify/assert"
)

type FakeFinder struct {
	Find func(giota.FindTransactionsRequest) ([]giota.Transaction, error)
	Read func([]giota.Trytes) ([]giota.Transaction, error)
}

func (f *FakeFinder) FindTransactions(req giota.FindTransactionsRequest) ([]giota.Transaction, error) {
	return f.Find(req)
}
func (f *FakeFinder) ReadTransactions(t []giota.Trytes) ([]giota.Transaction, error) {
	return f.Read(t)
}

func TestReadTransactions(t *testing.T) {
	assert := assert.New(t)

	account, err := giota.NewAddress(giota.NewSeed(), 0, 1)
	assert.Nil(err)

	txs, err := ReadTransactions(string(account), &FakeFinder{
		Find: func(req giota.FindTransactionsRequest) ([]giota.Transaction, error) {
			assert.Len(req.Addresses, 1)
			assert.EqualValues(account, req.Addresses[0])

			message, err := mamutils.ToMAMTrytes("Test")
			assert.Nil(err)
			return []giota.Transaction{
				giota.Transaction{
					SignatureMessageFragment: message,
					Value:     1000,
					Timestamp: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local),
					Address:   "Recipient",
				},
			}, nil
		},
	})

	assert.Nil(err)
	assert.Len(txs, 1)
	assert.EqualValues(Transaction{
		Message:   "Test",
		Value:     1000,
		Timestamp: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local),
		Recipient: "Recipient",
	}, txs[0])
}
func TestReadTransactionsError(t *testing.T) {
	assert := assert.New(t)

	txs, err := ReadTransactions("WRONGADRESS", &FakeFinder{})

	assert.NotNil(err)
	assert.Len(txs, 0)
}
func TestReadTransactionsErrorRequest(t *testing.T) {
	assert := assert.New(t)

	account, err := giota.NewAddress(giota.NewSeed(), 0, 1)
	assert.Nil(err)

	txs, err := ReadTransactions(string(account), &FakeFinder{
		Find: func(req giota.FindTransactionsRequest) ([]giota.Transaction, error) {
			assert.Len(req.Addresses, 1)
			assert.EqualValues(account, req.Addresses[0])

			return []giota.Transaction{}, fmt.Errorf("SOMEERROR")
		},
	})

	assert.NotNil(err)
	assert.Len(txs, 0)
}

func TestReadTransactionById(t *testing.T) {
	assert := assert.New(t)

	someTransactionID, err := giota.NewAddress(giota.NewSeed(), 0, 1)
	assert.Nil(err)

	tx, err := ReadTransaction(string(someTransactionID), &FakeFinder{
		Read: func(hashes []giota.Trytes) ([]giota.Transaction, error) {
			assert.Len(hashes, 1)
			assert.EqualValues(someTransactionID, hashes[0])

			message, err := mamutils.ToMAMTrytes("Test")
			assert.Nil(err)
			return []giota.Transaction{
				giota.Transaction{
					SignatureMessageFragment: message,
					Value:     1000,
					Timestamp: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local),
					Address:   "Recipient",
				},
			}, nil
		},
	})

	assert.Nil(err)
	assert.EqualValues(Transaction{
		Message:   "Test",
		Value:     1000,
		Timestamp: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local),
		Recipient: "Recipient",
	}, tx)
}

func TestReadTransactionByIdError(t *testing.T) {
	assert := assert.New(t)

	tx, err := ReadTransaction("WRONGID123", &FakeFinder{})

	assert.NotNil(err)
	assert.EqualValues(Transaction{}, tx)
}
func TestReadTransactionByIdErrorRequest(t *testing.T) {
	assert := assert.New(t)

	someTransactionID, err := giota.NewAddress(giota.NewSeed(), 0, 1)
	assert.Nil(err)

	tx, err := ReadTransaction(string(someTransactionID), &FakeFinder{
		Read: func(hashes []giota.Trytes) ([]giota.Transaction, error) {
			assert.Len(hashes, 1)
			assert.EqualValues(someTransactionID, hashes[0])

			return []giota.Transaction{}, fmt.Errorf("SOMEERROR")
		},
	})

	assert.NotNil(err)
	assert.EqualValues(Transaction{}, tx)
}
