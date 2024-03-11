package iotago

import (
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// ContextInputType defines the type of context inputs.
type ContextInputType byte

const (
	// ContextInputCommitment is a type of input which references a commitment.
	ContextInputCommitment ContextInputType = iota
	// ContextInputBlockIssuanceCredit is a type of input which references the block issuance credit from a specific account and commitment, the latter being provided by a commitment input.
	ContextInputBlockIssuanceCredit
	// ContextInputReward is a type of input which references an Account or Delegation Input for which to claim rewards.
	ContextInputReward
)

func (inputType ContextInputType) String() string {
	if int(inputType) >= len(contextInputNames) {
		return fmt.Sprintf("unknown input type: %d", inputType)
	}

	return contextInputNames[inputType]
}

var contextInputNames = [ContextInputReward + 1]string{"CommitmentInput", "BlockIssuanceCreditInput", "RewardInput"}

// ContextInput provides an additional contextual input for transaction validation.
type ContextInput interface {
	Sizer
	constraints.Cloneable[ContextInput]
	constraints.Comparable[ContextInput]
	ProcessableObject

	// Type returns the type of ContextInput.
	Type() ContextInputType
}

// ContextInputs is a slice of ContextInput.
type ContextInputs[T ContextInput] []T

func (in ContextInputs[T]) Clone() ContextInputs[T] {
	cpy := make(ContextInputs[T], len(in))
	for idx, input := range in {
		//nolint:forcetypeassert // we can safely assume that this is of type T
		cpy[idx] = input.Clone().(T)
	}

	return cpy
}

func (in ContextInputs[T]) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	var workScoreContextInputs WorkScore
	for _, input := range in {
		workScoreInput, err := input.WorkScore(workScoreParameters)
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

// Sort sorts the Context Inputs in lexical order.
func (in ContextInputs[T]) Sort() {
	sort.Slice(in, func(i, j int) bool {
		return in[i].Compare(in[j]) < 0
	})
}

// ContextInputsRewardInputMaxIndex returns a ElementValidationFunc
// which checks that every Reward Input references an index <= max inputs count.
func ContextInputsRewardInputMaxIndex(inputsCount uint16) ElementValidationFunc[ContextInput] {
	return func(index int, input ContextInput) error {
		switch castInput := input.(type) {
		case *CommitmentInput, *BlockIssuanceCreditInput:
		case *RewardInput:
			utxoIndex := castInput.Index
			if utxoIndex >= inputsCount {
				return ierrors.WithMessagef(ErrInputRewardIndexExceedsMaxInputsCount, "reward input %d references index %d which is equal or greater than the inputs count %d",
					index, utxoIndex, inputsCount)
			}
		default:
			panic("all supported context input types should be handled above")
		}

		return nil
	}
}

// ContextInputsCommitmentInputRequirement returns an ElementValidationFunc which
// checks that a Commitment Input is present if a BIC or Reward Input is present.
func ContextInputsCommitmentInputRequirement() ElementValidationFunc[ContextInput] {
	// Once we see the first BIC or Reward Input and there was no Commitment Input before, then due to lexical ordering
	// a commitment input cannot appear later, so we can error immediately.
	seenCommitmentInput := false

	return func(index int, input ContextInput) error {
		switch input.(type) {
		case *CommitmentInput:
			seenCommitmentInput = true
		case *BlockIssuanceCreditInput:
			if !seenCommitmentInput {
				return ierrors.WithMessagef(ErrCommitmentInputMissing, "block issuance credit input at index %d requires a commitment input", index)
			}
		case *RewardInput:
			if !seenCommitmentInput {
				return ierrors.WithMessagef(ErrCommitmentInputMissing, "reward input at index %d requires a commitment input", index)
			}
		default:
			panic("all supported context input types should be handled above")
		}

		return nil
	}
}
