package iotago

import (
	"fmt"

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

// ContextInputsSyntacticalValidationFunc which given the index of an input and the input itself,
// runs syntactical validations and returns an error if any should fail.
type ContextInputsSyntacticalValidationFunc func(index int, input ContextInput) error

// ContextInputsSyntacticalUnique returns a ContextInputsSyntacticalValidationFunc
// which checks that
//   - there are exactly 0 or 1 Commitment inputs.
//   - every Block Issuance Credits Input references a different account.
//   - every Reward Input references a different input and the index it references is <= max inputs count.
func ContextInputsSyntacticalUnique() ContextInputsSyntacticalValidationFunc {
	hasCommitment := false
	bicSet := map[string]int{}
	rewardSet := map[uint16]int{}

	return func(index int, input ContextInput) error {
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
