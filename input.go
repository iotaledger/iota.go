package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// InputType defines the type of inputs.
type InputType byte

const (
	// InputUTXO is a type of input which references an unspent transaction output.
	InputUTXO InputType = iota
	// InputTreasury is a type of input which references a milestone which generated a treasury output.
	InputTreasury
)

func (inputType InputType) String() string {
	if int(inputType) >= len(inputNames) {
		return fmt.Sprintf("unknown input type: %d", inputType)
	}
	return inputNames[inputType]
}

var (
	inputNames = [InputTreasury + 1]string{"UTXOInput", "TreasuryInput"}
)

var (
	// ErrRefUTXOIndexInvalid gets returned on invalid UTXO indices.
	ErrRefUTXOIndexInvalid = fmt.Errorf("the referenced UTXO index must be between %d and %d (inclusive)", RefUTXOIndexMin, RefUTXOIndexMax)
)

// Inputs a slice of Input.
type Inputs[T Input] []T

func (in Inputs[T]) Size() int {
	sum := serializer.UInt16ByteSize
	for _, i := range in {
		sum += i.Size()
	}
	return sum
}

// Input references a UTXO.
type Input interface {
	Sizer

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

// InputsSyntacticalValidationFunc which given the index of an input and the input itself, runs syntactical validations and returns an error if any should fail.
type InputsSyntacticalValidationFunc func(index int, input Input) error

// InputsSyntacticalUnique returns an InputsSyntacticalValidationFunc which checks that every input has a unique UTXO ref.
func InputsSyntacticalUnique() InputsSyntacticalValidationFunc {
	set := map[string]int{}
	return func(index int, input Input) error {
		ref, is := input.(IndexedUTXOReferencer)
		if !is {
			return fmt.Errorf("%w: input %d, tx can only contain IndexedUTXOReferencer inputs", ErrUnsupportedInputType, index)
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
			return fmt.Errorf("%w: input %d, tx can only contain IndexedUTXOReferencer inputs", ErrUnsupportedInputType, index)
		}
		if ref.Index() < RefUTXOIndexMin || ref.Index() > RefUTXOIndexMax {
			return fmt.Errorf("%w: input %d", ErrRefUTXOIndexInvalid, index)
		}
		return nil
	}
}

// SyntacticallyValidateInputs validates the inputs by running them against the given InputsSyntacticalValidationFunc(s).
func SyntacticallyValidateInputs(inputs Inputs[TxEssenceInput], funcs ...InputsSyntacticalValidationFunc) error {
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
