package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrArrayValidationOrderViolatesLexicalOrder gets returned if the array elements are not in lexical order.
	ErrArrayValidationOrderViolatesLexicalOrder = ierrors.New("array elements must be in their lexical order (byte wise)")
	// ErrArrayValidationViolatesUniqueness gets returned if the array elements are not unique.
	ErrArrayValidationViolatesUniqueness = ierrors.New("array elements must be unique")
)

// TODO
type ElementValidationFunc[T any] func(index int, next T) error

// TODO: Extend doc.
// Helper function to validate a slice syntactically.
func SyntacticSliceValidator[T constraints.Comparable[T]](ctx context.Context, slice []T, validationFunc ElementValidationFunc[T]) error {
	for i, element := range slice {
		if err := validationFunc(i, element); err != nil {
			return err
		}
	}

	return nil
}

// TODO
func LexicalOrderAndUniqueness[T constraints.Comparable[T]](slice []T) ElementValidationFunc[T] {
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
