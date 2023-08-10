package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// StateType defines the type of inputs.
type StateType byte

const (
	// InputUTXO is a type of input which references an unspent transaction output.
	InputUTXO StateType = iota
	// InputCommitment is a type of input which references a commitment.
	InputCommitment
	// InputBlockIssuanceCredit is a type of input which references the block issuance credit from a specific account and commitment, the latter being provided by a commitment input.
	InputBlockIssuanceCredit
	// InputReward is a type of input which references an Account or Delegation Input for which to claim rewards.
	InputReward
)

func (inputType StateType) String() string {
	if int(inputType) >= len(inputNames) {
		return fmt.Sprintf("unknown input type: %d", inputType)
	}

	return inputNames[inputType]
}

var inputNames = [InputUTXO + 1]string{"UTXOInput"}

var (
	// ErrRefUTXOIndexInvalid gets returned on invalid UTXO indices.
	ErrRefUTXOIndexInvalid = ierrors.Errorf("the referenced UTXO index must be between %d and %d (inclusive)", RefUTXOIndexMin, RefUTXOIndexMax)
	// ErrUnknownContextInputType gets returned for unknown context input types.
	ErrUnknownContextInputType = ierrors.New("unknown context input type")
)

// Inputs is a slice of Input.
type Inputs[T Input] []T

func (in Inputs[T]) Size() int {
	sum := serializer.UInt16ByteSize
	for _, i := range in {
		sum += i.Size()
	}

	return sum
}

func (in Inputs[T]) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// LengthPrefixType
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(serializer.UInt16ByteSize)
	if err != nil {
		return 0, err
	}

	for _, input := range in {
		workScoreInput, err := input.WorkScore(workScoreStructure)
		if err != nil {
			return 0, err
		}

		workScoreBytes, err = workScoreBytes.Add(workScoreInput)
		if err != nil {
			return 0, err
		}
	}

	return workScoreBytes, nil
}

// Input references a generic input.
type Input interface {
	Sizer
	ProcessableObject

	StateID() Identifier

	// Type returns the type of Input.
	Type() StateType
}

// InputsSyntacticalValidationFunc which given the index of an input and the input itself, runs syntactical validations and returns an error if any should fail.
type InputsSyntacticalValidationFunc func(index int, input Input) error

// InputsSyntacticalUnique returns an InputsSyntacticalValidationFunc which checks that every input has a unique UTXO ref.
func InputsSyntacticalUnique() InputsSyntacticalValidationFunc {
	utxoSet := map[string]int{}

	return func(index int, input Input) error {
		switch castInput := input.(type) {
		case *UTXOInput:
			utxoRef := castInput.OutputID()
			k := string(utxoRef[:])
			if j, has := utxoSet[k]; has {
				return ierrors.Wrapf(ErrInputUTXORefsNotUnique, "input %d and %d share the same UTXO ref", j, index)
			}
			utxoSet[k] = index
		default:
			return ierrors.Wrapf(ErrUnknownInputType, "input %d, tx can only contain UTXO inputs", index)
		}

		return nil
	}
}

// InputsSyntacticalIndicesWithinBounds returns an InputsSyntacticalValidationFunc which checks that the UTXO ref index is within bounds.
func InputsSyntacticalIndicesWithinBounds() InputsSyntacticalValidationFunc {
	return func(index int, input Input) error {
		switch castInput := input.(type) {
		case *UTXOInput:
			if castInput.Index() < RefUTXOIndexMin || castInput.Index() > RefUTXOIndexMax {
				return ierrors.Wrapf(ErrRefUTXOIndexInvalid, "input %d", index)
			}
		default:
			return ierrors.Wrapf(ErrUnknownInputType, "input %d, tx can only contain UTXInput inputs", index)
		}

		return nil
	}
}

// SyntacticallyValidateInputs validates the inputs by running them against the given InputsSyntacticalValidationFunc(s).
func SyntacticallyValidateInputs(inputs TxEssenceInputs, funcs ...InputsSyntacticalValidationFunc) error {
	for i, input := range inputs {
		for _, f := range funcs {
			if err := f(i, input); err != nil {
				return err
			}
		}
	}

	return nil
}
