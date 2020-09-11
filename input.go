package iota

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// Defines the type of inputs.
type InputType = byte

const (
	// A type of input which references an unspent transaction output.
	InputUTXO InputType = iota

	// The minimum index of a referenced UTXO.
	RefUTXOIndexMin = 0
	// The maximum index of a referenced UTXO.
	RefUTXOIndexMax = 126

	// The size of a UTXO input: input type + tx id + index
	UTXOInputSize = SmallTypeDenotationByteSize + TransactionIDLength + UInt16ByteSize
)

var (
	// Returned on invalid UTXO indices.
	ErrRefUTXOIndexInvalid = fmt.Errorf("the referenced UTXO index must be between %d and %d (inclusive)", RefUTXOIndexMin, RefUTXOIndexMax)
)

// InputSelector implements SerializableSelectorFunc for input types.
func InputSelector(inputType uint32) (Serializable, error) {
	var seri Serializable
	switch byte(inputType) {
	case InputUTXO:
		seri = &UTXOInput{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownInputType, inputType)
	}
	return seri, nil
}

// UTXOInputID defines the identifier for an UTXO input which consists
// out of the referenced transaction hash and the given output index.
type UTXOInputID [TransactionIDLength + UInt16ByteSize]byte

// UTXOInput references an unspent transaction output by the signed transaction payload's hash and the corresponding index of the output.
type UTXOInput struct {
	// The transaction ID of the referenced transaction.
	TransactionID [TransactionIDLength]byte `json:"transaction_id"`
	// The output index of the output on the referenced transaction.
	TransactionOutputIndex uint16 `json:"transaction_output_index"`
}

// ID returns the UTXOInputID.
func (u *UTXOInput) ID() UTXOInputID {
	var id UTXOInputID
	copy(id[:TransactionIDLength], u.TransactionID[:])
	binary.LittleEndian.PutUint16(id[TransactionIDLength:], u.TransactionOutputIndex)
	return id
}

func (u *UTXOInput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(UTXOInputSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid UTXO input bytes: %w", err)
		}
		if err := checkTypeByte(data, InputUTXO); err != nil {
			return 0, fmt.Errorf("unable to deserialize UTXO input: %w", err)
		}
	}

	data = data[SmallTypeDenotationByteSize:]

	// read transaction id
	copy(u.TransactionID[:], data[:TransactionIDLength])
	data = data[TransactionIDLength:]

	// output index
	u.TransactionOutputIndex = binary.LittleEndian.Uint16(data)

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := utxoInputRefBoundsValidator(-1, u); err != nil {
			return 0, fmt.Errorf("%w: unable to deserialize UTXO input", err)
		}
	}

	return UTXOInputSize, nil
}

func (u *UTXOInput) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := utxoInputRefBoundsValidator(-1, u); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize UTXO input", err)
		}
	}

	var b [UTXOInputSize]byte
	b[0] = InputUTXO
	copy(b[SmallTypeDenotationByteSize:], u.TransactionID[:])
	binary.LittleEndian.PutUint16(b[UTXOInputSize-UInt16ByteSize:], u.TransactionOutputIndex)
	return b[:], nil
}

// InputsValidatorFunc which given the index of an input and the input itself, runs validations and returns an error if any should fail.
type InputsValidatorFunc func(index int, input *UTXOInput) error

// InputsUTXORefsUniqueValidator returns a validator which checks that every input has a unique UTXO ref.
func InputsUTXORefsUniqueValidator() InputsValidatorFunc {
	set := map[string]int{}
	return func(index int, input *UTXOInput) error {
		var b strings.Builder
		if _, err := b.Write(input.TransactionID[:]); err != nil {
			return fmt.Errorf("%w: unable to write tx ID in ref validator", err)
		}
		if err := binary.Write(&b, binary.LittleEndian, input.TransactionOutputIndex); err != nil {
			return fmt.Errorf("%w: unable to write UTXO index in ref validator", err)
		}
		k := b.String()
		if j, has := set[k]; has {
			return fmt.Errorf("%w: input %d and %d share the same UTXO ref", ErrInputUTXORefsNotUnique, j, index)
		}
		set[k] = index
		return nil
	}
}

// InputsUTXORefIndexBoundsValidator returns a validator which checks that the UTXO ref index is within bounds.
func InputsUTXORefIndexBoundsValidator() InputsValidatorFunc {
	return func(index int, input *UTXOInput) error {
		if input.TransactionOutputIndex < RefUTXOIndexMin || input.TransactionOutputIndex > RefUTXOIndexMax {
			return fmt.Errorf("%w: input %d", ErrRefUTXOIndexInvalid, index)
		}
		return nil
	}
}

var utxoInputRefBoundsValidator = InputsUTXORefIndexBoundsValidator()

// ValidateInputs validates the inputs by running them against the given InputsValidatorFunc.
func ValidateInputs(inputs Serializables, funcs ...InputsValidatorFunc) error {
	for i, input := range inputs {
		dep, ok := input.(*UTXOInput)
		if !ok {
			return fmt.Errorf("%w: can only validate on UTXO inputs", ErrUnknownInputType)
		}
		for _, f := range funcs {
			if err := f(i, dep); err != nil {
				return err
			}
		}
	}
	return nil
}
