package iotago

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/serializer/v2"
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
	TransactionID TransactionID `serix:""`
	// The output index of the output on the referenced transaction.
	TransactionOutputIndex uint16 `serix:""`
}

func (u *UTXOInput) Clone() Input {
	return &UTXOInput{
		TransactionID:          u.TransactionID,
		TransactionOutputIndex: u.TransactionOutputIndex,
	}
}

func (u *UTXOInput) UniquenessIdentifier() []byte {
	id := make([]byte, 0, u.Size())

	id = append(id, byte(u.Type()))
	id = append(id, u.TransactionID[:]...)
	id = binary.LittleEndian.AppendUint16(id, u.TransactionOutputIndex)

	return id
}

func (u *UTXOInput) Type() InputType {
	return InputUTXO
}

func (u *UTXOInput) IsReadOnly() bool {
	return false
}

func (u *UTXOInput) OutputID() OutputID {
	var id OutputID
	copy(id[:TransactionIDLength], u.TransactionID[:])
	binary.LittleEndian.PutUint16(id[TransactionIDLength:], u.TransactionOutputIndex)

	return id
}

func (u *UTXOInput) Index() uint16 {
	return u.TransactionOutputIndex
}

// CreationSlot returns the slot the Output was created in.
func (u *UTXOInput) CreationSlot() SlotIndex {
	return u.TransactionID.Slot()
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
	// InputType + TransactionID + TransactionOutputIndex
	return serializer.SmallTypeDenotationByteSize + TransactionIDLength + OutputIndexLength
}

func (u *UTXOInput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// inputs require lookup of the UTXO, so requires extra work.
	return workScoreParameters.Input, nil
}
