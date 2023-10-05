package iotago

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

func EmptyOutputIDWithCreationSlot(slot SlotIndex) OutputID {
	var outputID OutputID
	binary.LittleEndian.PutUint32(outputID[IdentifierLength:OutputIDLength], uint32(slot))
	return outputID
}

// TransactionID returns the TransactionID of the Output this OutputID references.
func (o OutputID) TransactionID() TransactionID {
	var txID TransactionID
	copy(txID[:], o[:TransactionIDLength])

	return txID
}

// CreationSlot returns the slot the Output was created in.
func (o OutputID) CreationSlot() SlotIndex {
	return o.TransactionID().Slot()
}

// UTXOInput creates a UTXOInput from this OutputID.
func (o OutputID) UTXOInput() *UTXOInput {
	return &UTXOInput{
		TransactionID:          o.TransactionID(),
		TransactionOutputIndex: o.Index(),
	}
}

// OutputIDFromTransactionIDAndIndex creates a OutputID from the given TransactionID and output index.
func OutputIDFromTransactionIDAndIndex(txID TransactionID, index uint16) OutputID {
	utxo := &UTXOInput{
		TransactionID:          txID,
		TransactionOutputIndex: index,
	}

	return utxo.OutputID()
}

// UTXOInputs converts the OutputIDs slice to Inputs.
func (ids OutputIDs) UTXOInputs() TxEssenceInputs {
	inputs := make(TxEssenceInputs, 0)
	for _, outputID := range ids {
		inputs = append(inputs, outputID.UTXOInput())
	}

	return inputs
}

// OrderedSet returns an Outputs slice ordered by this OutputIDs slice given an OutputSet.
func (ids OutputIDs) OrderedSet(set OutputSet) Outputs[Output] {
	outputs := make(Outputs[Output], len(ids))
	for i, outputID := range ids {
		outputs[i] = set[outputID]
	}

	return outputs
}

// OutputIDHex is the hex representation of an output ID.
type OutputIDHex string

// MustSplitParts returns the transaction ID and output index parts of the hex output ID.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustSplitParts() (*TransactionID, uint16) {
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		panic(err)
	}

	return txID, outputIndex
}

// SplitParts returns the transaction ID and output index parts of the hex output ID.
func (oih OutputIDHex) SplitParts() (*TransactionID, uint16, error) {
	outputIDBytes, err := hexutil.DecodeHex(string(oih))
	if err != nil {
		return nil, 0, err
	}
	var txID TransactionID
	copy(txID[:], outputIDBytes[:TransactionIDLength])
	outputIndex := binary.LittleEndian.Uint16(outputIDBytes[TransactionIDLength : TransactionIDLength+serializer.UInt16ByteSize])

	return &txID, outputIndex, nil
}

// MustAsUTXOInput converts the hex output ID to a UTXOInput.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustAsUTXOInput() *UTXOInput {
	utxoInput, err := oih.AsUTXOInput()
	if err != nil {
		panic(err)
	}

	return utxoInput
}

// AsUTXOInput converts the hex output ID to a UTXOInput.
func (oih OutputIDHex) AsUTXOInput() (*UTXOInput, error) {
	var utxoInput UTXOInput
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		return nil, err
	}
	copy(utxoInput.TransactionID[:], txID[:])
	utxoInput.TransactionOutputIndex = outputIndex

	return &utxoInput, nil
}

// HexOutputIDs is a slice of hex encoded OutputID strings.
type HexOutputIDs []string

// MustOutputIDs converts the hex strings into OutputIDs.
func (ids HexOutputIDs) MustOutputIDs() OutputIDs {
	vals, err := ids.OutputIDs()
	if err != nil {
		panic(err)
	}

	return vals
}

// OutputIDs converts the hex strings into OutputIDs.
func (ids HexOutputIDs) OutputIDs() (OutputIDs, error) {
	vals := make(OutputIDs, len(ids))
	for i, v := range ids {
		val, err := hexutil.DecodeHex(v)
		if err != nil {
			return nil, err
		}
		copy(vals[i][:], val)
	}

	return vals, nil
}
