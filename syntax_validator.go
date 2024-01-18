package iotago

import (
	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
)

// ElementValidationFunc is a func that, given the index of a slice element and the element itself
// runs syntactical validations and returns an error if it fails.
type ElementValidationFunc[T any] func(index int, next T) error

// An ElementValidationFunc that checks lexical order and uniqueness based on the Compare implementation.
func LexicalOrderAndUniquenessValidator[T constraints.Comparable[T]]() ElementValidationFunc[T] {
	var prev *T
	var prevIndex int

	return func(index int, next T) error {
		if prev == nil {
			prev = &next
			prevIndex = index
		} else {
			switch (*prev).Compare(next) {
			case 1:
				return ierrors.Wrapf(ErrArrayValidationOrderViolatesLexicalOrder, "element %d should have been before element %d", index, prevIndex)
			case 0:
				return ierrors.Wrapf(ErrArrayValidationViolatesUniqueness, "element %d and element %d are duplicates", index, prevIndex)
			}

			prev = &next
			prevIndex = index
		}

		return nil
	}
}

// SyntacticSliceValidator iterates over a slice and calls elementValidationFunc on each element,
// returning the first error it encounters, if any.
func SyntacticSliceValidator[T any](
	slice []T,
	elementValidationFunc ElementValidationFunc[T],
) error {
	for i, element := range slice {
		if err := elementValidationFunc(i, element); err != nil {
			return err
		}
	}

	return nil
}
