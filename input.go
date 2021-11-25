package iotago

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/v2/serializer"
)

// InputType defines the type of inputs.
type InputType byte

const (
	// InputUTXO is a type of input which references an unspent transaction output.
	InputUTXO InputType = iota
	// InputTreasury is a type of input which references a milestone which generated a treasury output.
	InputTreasury
)

var (
	// ErrRefUTXOIndexInvalid gets returned on invalid UTXO indices.
	ErrRefUTXOIndexInvalid = fmt.Errorf("the referenced UTXO index must be between %d and %d (inclusive)", RefUTXOIndexMin, RefUTXOIndexMax)

	// ErrTypeIsNotSupportedInput gets returned when a serializable was found to not be a supported Input.
	ErrTypeIsNotSupportedInput = errors.New("serializable is not a supported input")
)

// Inputs a slice of Input.
type Inputs []Input

func (in Inputs) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(in))
	for i, x := range in {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (in *Inputs) FromSerializables(seris serializer.Serializables) {
	*in = make(Inputs, len(seris))
	for i, seri := range seris {
		(*in)[i] = seri.(Input)
	}
}

// Input references a UTXO.
type Input interface {
	serializer.Serializable

	// Type returns the type of Input.
	Type() InputType
}

// IndexedUTXOReferencer is a type of Input which references a UTXO by the transaction ID and output index.
type IndexedUTXOReferencer interface {
	Input

	// Ref returns the UTXO this Input references.
	Ref() OutputID
	// Index returns the output index of the UTXO this Input references.
	Index() uint16
}

// InputTypeToString returns the name of an Input given the type.
func InputTypeToString(ty InputType) string {
	switch ty {
	case InputUTXO:
		return "UTXOInput"
	case InputTreasury:
		return "TreasuryInput"
	default:
		return "unknown input"
	}
}

// InputSelector implements SerializableSelectorFunc for input types.
func InputSelector(inputType uint32) (Input, error) {
	var seri Input
	switch InputType(inputType) {
	case InputUTXO:
		seri = &UTXOInput{}
	case InputTreasury:
		seri = &TreasuryInput{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownInputType, inputType)
	}
	return seri, nil
}

// InputsSyntacticalValidationFunc which given the index of an input and the input itself, runs syntactical validations and returns an error if any should fail.
type InputsSyntacticalValidationFunc func(index int, input Input) error

// InputsSyntacticalUnique returns an InputsSyntacticalValidationFunc which checks that every input has a unique UTXO ref.
func InputsSyntacticalUnique() InputsSyntacticalValidationFunc {
	set := map[string]int{}
	return func(index int, input Input) error {
		ref, is := input.(IndexedUTXOReferencer)
		if !is {
			return nil
		}
		utxoRef := ref.Ref()
		k := string(utxoRef[:])
		if j, has := set[k]; has {
			return fmt.Errorf("%w: input %d and %d share the same UTXO ref", ErrInputUTXORefsNotUnique, j, index)
		}
		set[k] = index
		return nil
	}
}

// InputsSyntacticalIndicesWithinBounds returns an InputsSyntacticalValidationFunc which checks that the UTXO ref index is within bounds.
func InputsSyntacticalIndicesWithinBounds() InputsSyntacticalValidationFunc {
	return func(index int, input Input) error {
		ref, is := input.(IndexedUTXOReferencer)
		if !is {
			return nil
		}
		if ref.Index() < RefUTXOIndexMin || ref.Index() > RefUTXOIndexMax {
			return fmt.Errorf("%w: input %d", ErrRefUTXOIndexInvalid, index)
		}
		return nil
	}
}

var inputsPredicateIndicesWithinBounds = InputsSyntacticalIndicesWithinBounds()

// ValidateInputs validates the inputs by running them against the given InputsSyntacticalValidationFunc(s).
func ValidateInputs(inputs Inputs, funcs ...InputsSyntacticalValidationFunc) error {
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
	switch InputType(ty) {
	case InputUTXO:
		obj = &jsonUTXOInput{}
	case InputTreasury:
		obj = &jsonTreasuryInput{}
	default:
		return nil, fmt.Errorf("unable to decode input type from JSON: %w", ErrUnknownInputType)
	}
	return obj, nil
}
