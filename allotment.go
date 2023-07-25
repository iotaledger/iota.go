package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

// BlockIssuanceCredits defines the type of block issuance credits.
type BlockIssuanceCredits int64

// Allotments is a slice of Allotment.
type Allotments []*Allotment

// Allotment is a struct that represents a list of account IDs and an allotted value.
type Allotment struct {
	AccountID AccountID `serix:"0"`
	Value     Mana      `serix:"1"`
}

func (a Allotments) Size() int {
	return len(a) * (AccountIDLength + ManaSize)
}

func (a Allotments) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(a.Size())
	if err != nil {
		return 0, err
	}

	// Allotments requires invocation of account managers, so requires extra work.
	workScoreAllotments, err := workScoreStructure.Allotment.Multiply(len(a))
	if err != nil {
		return 0, err
	}

	return workScoreBytes.Add(workScoreAllotments)
}

func (a Allotments) Get(id AccountID) Mana {
	for _, allotment := range a {
		if allotment.AccountID == id {
			return allotment.Value
		}
	}
	return 0
}

// AllotmentsSyntacticalValidationFunc which given the index of an Allotment and the Allotment itself, runs syntactical validations and returns an error if any should fail.
type AllotmentsSyntacticalValidationFunc func(index int, input *Allotment) error

// SyntacticallyValidateAllotments validates the allotments by running them against the given AllotmentsSyntacticalValidationFunc(s).
func SyntacticallyValidateAllotments(allotments TxEssenceAllotments, funcs ...AllotmentsSyntacticalValidationFunc) error {
	for i, allotment := range allotments {
		for _, f := range funcs {
			if err := f(i, allotment); err != nil {
				return err
			}
		}
	}

	return nil
}

// AllotmentsSyntacticalUnique returns an AllotmentsSyntacticalValidationFunc which checks that every Allotment has a unique Account.
func AllotmentsSyntacticalUnique() AllotmentsSyntacticalValidationFunc {
	allotmentsSet := map[string]int{}
	return func(index int, allotment *Allotment) error {
		k := string(allotment.AccountID[:])
		if j, has := allotmentsSet[k]; has {
			return ierrors.Wrapf(ErrAllotmentsNotUnique, "allotment %d and %d share the same Account", j, index)
		}
		allotmentsSet[k] = index

		return nil
	}
}
