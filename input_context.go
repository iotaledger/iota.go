package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// ContextInputType defines the type of context inputs.
type ContextInputType byte

func (inputType ContextInputType) String() string {
	if int(inputType) >= len(contextInputNames) {
		return fmt.Sprintf("unknown input type: %d", inputType)
	}

	return contextInputNames[inputType]
}

var contextInputNames = [InputReward + 1]string{"CommitmentInput", "BlockIssuanceCreditInput", "RewardInput"}

// ContextInputs is a slice of ContextInput.
type ContextInputs[T Input] []T

func (in ContextInputs[T]) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	var workScoreContextInputs WorkScore
	for _, input := range in {
		workScoreInput, err := input.WorkScore(workScoreStructure)
		if err != nil {
			return 0, err
		}

		workScoreContextInputs, err = workScoreContextInputs.Add(workScoreInput)
		if err != nil {
			return 0, err
		}
	}

	return workScoreContextInputs, nil
}

func (in ContextInputs[T]) Size() int {
	sum := serializer.UInt16ByteSize
	for _, i := range in {
		sum += i.Size()
	}

	return sum
}

// ContextInputsSyntacticalValidationFunc which given the index of an input and the input itself,
// runs syntactical validations and returns an error if any should fail.
type ContextInputsSyntacticalValidationFunc func(index int, input Input) error

// ContextInputsSyntacticalUnique returns a ContextInputsSyntacticalValidationFunc
// which checks that
//   - there are exactly 0 or 1 Commitment inputs.
//   - every Block Issuance Credits Input references a different account.
//   - every Reward Input references a different input and the index it references is <= max inputs count.
func ContextInputsSyntacticalUnique() ContextInputsSyntacticalValidationFunc {
	hasCommitment := false
	bicSet := map[string]int{}
	rewardSet := map[uint16]int{}

	return func(index int, input Input) error {
		switch castInput := input.(type) {
		case *BlockIssuanceCreditInput:
			accountID := castInput.AccountID
			k := string(accountID[:])
			if j, has := bicSet[k]; has {
				return ierrors.Wrapf(ErrInputBICNotUnique, "input %d and %d share the same Account ref", j, index)
			}
			bicSet[k] = index
		case *RewardInput:
			utxoIndex := castInput.Index
			if utxoIndex > MaxInputsCount {
				return ierrors.Wrapf(ErrInputRewardInvalid, "input %d references an index greater than max inputs count", index)
			}
			if j, has := rewardSet[utxoIndex]; has {
				return ierrors.Wrapf(ErrInputRewardInvalid, "input %d and %d share the same input index", j, index)
			}
			rewardSet[utxoIndex] = index
		case *CommitmentInput:
			if hasCommitment {
				return ierrors.Wrapf(ErrMultipleInputCommitments, "input %d is the second commitment input", index)
			}
			hasCommitment = true
		case *UTXOInput:
			// ignore as we are evaluating context inputs only
		default:
			return ierrors.Wrapf(ErrUnknownContextInputType, "context input %d, tx can only contain CommitmentInputs, BlockIssuanceCreditInputs or RewardInputs", index)
		}

		return nil
	}
}

// SyntacticallyValidateContextInputs validates the context inputs by running them against
// the given ContextInputsSyntacticalValidationFunc(s).
func SyntacticallyValidateContextInputs(inputs TxEssenceContextInputs, funcs ...ContextInputsSyntacticalValidationFunc) error {
	for i, input := range inputs {
		for _, f := range funcs {
			if err := f(i, input); err != nil {
				return err
			}
		}
	}

	return nil
}
