package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// InputType defines the type of inputs.
type InputType byte

const (
	// InputUTXO is a type of input which references an unspent transaction output.
	InputUTXO InputType = iota
)

func (inputType InputType) String() string {
	if int(inputType) >= len(inputNames) {
		return fmt.Sprintf("unknown input type: %d", inputType)
	}

	return inputNames[inputType]
}

var inputNames = [InputUTXO + 1]string{"UTXOInput"}

var (
	// ErrRefUTXOIndexInvalid gets returned on invalid UTXO indices.
	ErrRefUTXOIndexInvalid = ierrors.Errorf("the referenced UTXO index must be between %d and %d (inclusive)", RefUTXOIndexMin, RefUTXOIndexMax)
)

// Inputs is a slice of Input.
type Inputs[T Input] []T

func (in Inputs[T]) Clone() Inputs[T] {
	cpy := make(Inputs[T], len(in))
	for idx, input := range in {
		//nolint:forcetypeassert // we can safely assume that this is of type T
		cpy[idx] = input.Clone().(T)
	}

	return cpy
}

func (in Inputs[T]) Size() int {
	sum := serializer.UInt16ByteSize
	for _, i := range in {
		sum += i.Size()
	}

	return sum
}

func (in Inputs[T]) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	var workScoreInputs WorkScore
	for _, input := range in {
		workScoreInput, err := input.WorkScore(workScoreParameters)
		if err != nil {
			return 0, err
		}

		workScoreInputs, err = workScoreInputs.Add(workScoreInput)
		if err != nil {
			return 0, err
		}
	}

	return workScoreInputs, nil
}

// Input references a generic input.
type Input interface {
	Sizer
	constraints.Cloneable[Input]
	ProcessableObject

	// Type returns the type of Input.
	Type() InputType
}

// InputsSyntacticalUnique returns an ElementValidationFunc which checks that every input has a unique reference UTXO index.
func InputsSyntacticalUnique() ElementValidationFunc[Input] {
	utxoSet := map[string]int{}

	return func(index int, input Input) error {
		switch castInput := input.(type) {
		case *UTXOInput:
			referencedOutputID := castInput.OutputID()
			k := string(referencedOutputID[:])
			if j, has := utxoSet[k]; has {
				return ierrors.WithMessagef(ErrInputUTXORefsNotUnique, "input %d and %d share the same referenced UTXO index", j, index)
			}
			utxoSet[k] = index
		default:
			panic("all supported input types should be handled above")
		}

		return nil
	}
}

// InputsSyntacticalIndicesWithinBounds returns an ElementValidationFunc which checks that the UTXO ref index is within bounds.
func InputsSyntacticalIndicesWithinBounds() ElementValidationFunc[Input] {
	return func(index int, input Input) error {
		switch castInput := input.(type) {
		case *UTXOInput:
			// TODO: do we really want to check the max value on the input side?
			if castInput.Index() < RefUTXOIndexMin || castInput.Index() > RefUTXOIndexMax {
				return ierrors.WithMessagef(ErrRefUTXOIndexInvalid, "input %d", index)
			}
		default:
			panic("all supported input types should be handled above")
		}

		return nil
	}
}
