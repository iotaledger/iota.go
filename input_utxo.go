package iotago

import (
	"encoding/binary"

	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// RefUTXOIndexMin is the minimum index of a referenced UTXO.
	RefUTXOIndexMin = 0
	// RefUTXOIndexMax is the maximum index of a referenced UTXO.
	RefUTXOIndexMax = MaxOutputsCount - 1
)

// UTXOInput references an unspent transaction output by the Transaction's ID and the corresponding index of the Output.
type UTXOInput struct {
	// The transaction ID of the referenced transaction.
	TransactionID TransactionID `serix:"0,mapKey=transactionId"`
	// The output index of the output on the referenced transaction.
	TransactionOutputIndex uint16 `serix:"1,mapKey=transactionOutputIndex"`
}

func (u *UTXOInput) Type() InputType {
	return InputUTXO
}

func (u *UTXOInput) Ref() OutputID {
	return u.ID()
}

func (u *UTXOInput) Index() uint16 {
	return u.TransactionOutputIndex
}

// ID returns the OutputID.
func (u *UTXOInput) ID() OutputID {
	var id OutputID
	copy(id[:TransactionIDLength], u.TransactionID[:])
	binary.LittleEndian.PutUint16(id[TransactionIDLength:], u.TransactionOutputIndex)
	return id
}

func (u *UTXOInput) Equals(other *UTXOInput) bool {
	if u == nil {
		return other == nil
	}
	if other == nil {
		return false
	}
	if u.TransactionID != other.TransactionID {
		return false
	}
	return u.TransactionOutputIndex == other.TransactionOutputIndex
}

func (u *UTXOInput) Size() int {
	return util.NumByteLen(byte(InputUTXO)) + TransactionIDLength + util.NumByteLen(u.TransactionOutputIndex)
}
