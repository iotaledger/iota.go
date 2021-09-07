package iotago

import (
	"encoding/binary"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"
	"strings"
)

// InputType defines the type of inputs.
type InputType = byte

const (
	// InputUTXO is a type of input which references an unspent transaction output.
	InputUTXO InputType = iota
	// InputTreasury is a type of input which references a milestone which generated a treasury output.
	InputTreasury
)

var (
	// ErrRefUTXOIndexInvalid gets returned on invalid UTXO indices.
	ErrRefUTXOIndexInvalid = fmt.Errorf("the referenced UTXO index must be between %d and %d (inclusive)", RefUTXOIndexMin, RefUTXOIndexMax)
)

// InputSelector implements SerializableSelectorFunc for input types.
func InputSelector(inputType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
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
func ValidateInputs(inputs serializer.Serializables, funcs ...InputsValidatorFunc) error {
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

// jsonInputSelector selects the json input implementation for the given type.
func jsonInputSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case InputUTXO:
		obj = &jsonUTXOInput{}
	case InputTreasury:
		obj = &jsonTreasuryInput{}
	default:
		return nil, fmt.Errorf("unable to decode input type from JSON: %w", ErrUnknownInputType)
	}
	return obj, nil
}
