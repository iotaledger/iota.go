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

var (
	// ErrUnknownContextInputType gets returned for unknown context input types.
	ErrUnknownContextInputType = ierrors.New("unknown context input type")
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
type ContextInputs[T ContextInput] []ContextInput

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

// ContextInputsSyntacticalLexicalOrderAndUniqueness returns a ElementValidationFunc
// which checks lexcial order and uniqueness.
//
// As a special case, it also checks that at most one commitment input is present,
// due to how Compare is defined on commitment inputs.
func ContextInputsSyntacticalLexicalOrderAndUniqueness() ElementValidationFunc[ContextInput] {
	contextInputValidationFunc := LexicalOrderAndUniquenessValidator[ContextInput]()
	return func(index int, input ContextInput) error {
		if err := contextInputValidationFunc(index, input); err != nil {
			return err
		}
		return nil
	}
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
				return ierrors.Wrapf(ErrInputRewardIndexExceedsMaxInputsCount, "reward input %d references index %d which is equal or greater than the inputs count %d",
					index, utxoIndex, inputsCount)
			}
		default:
			return ierrors.Wrapf(ErrUnknownContextInputType, "context input %d, tx can only contain CommitmentInputs, BlockIssuanceCreditInputs or RewardInputs", index)
		}

		return nil
	}
}
