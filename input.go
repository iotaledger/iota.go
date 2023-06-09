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
	// InputCommitment is a tupe of input which references a commitment.
	InputCommitment
	// InputBlockIssuanceCredit is a type of input which references the block issuance credit from a specific account and commitment.
	InputBlockIssuanceCredit
)

func (inputType InputType) String() string {
	if int(inputType) >= len(inputNames) {
		return fmt.Sprintf("unknown input type: %d", inputType)
	}
	return inputNames[inputType]
}

var (
	inputNames = [InputBlockIssuanceCredit + 1]string{"UTXOInput", "CommitmentInput", "BICInput"}
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
	utxoSet := map[string]int{}
	commitmentsSet := map[string]int{}
	bicSet := map[string]int{}
	return func(index int, input Input) error {
		switch castInput := input.(type) {
		case *BICInput:
			accountID := castInput.AccountID
			k := string(accountID[:])
			if j, has := utxoSet[k]; has {
				return fmt.Errorf("%w: input %d and %d share the same Account ref", ErrInputBICNotUnique, j, index)
			}
			utxoSet[k] = index
		case *CommitmentInput:
			commitmentID := castInput.CommitmentID
			k := string(commitmentID[:])
			if j, has := commitmentsSet[k]; has {
				return fmt.Errorf("%w: input %d and %d share the same Commitment ref", ErrInputCommitmentNotUnique, j, index)
			}
			commitmentsSet[k] = index
		case IndexedUTXOReferencer:
			utxoRef := castInput.Ref()
			k := string(utxoRef[:])
			if j, has := bicSet[k]; has {
				return fmt.Errorf("%w: input %d and %d share the same UTXO ref", ErrInputUTXORefsNotUnique, j, index)
			}
			bicSet[k] = index
		default:
			return fmt.Errorf("%w: input %d, tx can only contain IndexedUTXOReferencer, CommitmentInput or BICInput inputs", ErrUnsupportedInputType, index)
		}

		return nil
	}
}

// InputsSyntacticalIndicesWithinBounds returns an InputsSyntacticalValidationFunc which checks that the UTXO ref index is within bounds.
func InputsSyntacticalIndicesWithinBounds() InputsSyntacticalValidationFunc {
	return func(index int, input Input) error {
		switch castInput := input.(type) {
		case *BICInput:
			// TODO: any checks necessary?
		case *CommitmentInput:
			// TODO: any checks necessary?
		case IndexedUTXOReferencer:
			if castInput.Index() < RefUTXOIndexMin || castInput.Index() > RefUTXOIndexMax {
				return fmt.Errorf("%w: input %d", ErrRefUTXOIndexInvalid, index)
			}
		default:
			return fmt.Errorf("%w: input %d, tx can only contain IndexedUTXOReferencer, CommitmentInput or BICInput inputs", ErrUnsupportedInputType, index)
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

// TODO: does this belong here?
type ResolvedInputs struct {
	InputSet
	BICInputSet
	CommitmentInputSet
}
