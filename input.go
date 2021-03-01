package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// Defines the type of inputs.
type InputType = byte

const (
	// A type of input which references an unspent transaction output.
	InputUTXO InputType = iota
	// A type of input which references a milestone which generated a treasury output.
	InputTreasury
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
	case InputTreasury:
		seri = &TreasuryInput{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownInputType, inputType)
	}
	return seri, nil
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

// jsoninputselector selects the json input implementation for the given type.
func jsoninputselector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case InputUTXO:
		obj = &jsonutxoinput{}
	case InputTreasury:
		obj = &jsontreasuryinput{}
	default:
		return nil, fmt.Errorf("unable to decode input type from JSON: %w", ErrUnknownInputType)
	}
	return obj, nil
}

// jsonutxoinput defines the JSON representation of a UTXOInput.
type jsonutxoinput struct {
	Type                   int    `json:"type"`
	TransactionID          string `json:"transactionId"`
	TransactionOutputIndex int    `json:"transactionOutputIndex"`
}

func (j *jsonutxoinput) ToSerializable() (Serializable, error) {
	utxoInput := &UTXOInput{
		TransactionID:          [32]byte{},
		TransactionOutputIndex: uint16(j.TransactionOutputIndex),
	}
	transactionIDBytes, err := hex.DecodeString(j.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("unable to decode transaction ID from JSON for UTXO input: %w", err)
	}
	copy(utxoInput.TransactionID[:], transactionIDBytes)
	return utxoInput, nil
}
