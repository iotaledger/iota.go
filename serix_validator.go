package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrArrayValidationOrderViolatesLexicalOrder gets returned if the array elements are not in lexical order.
	ErrArrayValidationOrderViolatesLexicalOrder = ierrors.New("array elements must be in their lexical order (byte wise)")
	// ErrArrayValidationViolatesUniqueness gets returned if the array elements are not unique.
	ErrArrayValidationViolatesUniqueness = ierrors.New("array elements must be unique")
)

type LexicallyComparable[T any] interface {
	LexicalCompare(a T, b T) int
}

type ElementValidationFunc[T any] func(index int, next T) error

func LexicalOrderAndUniqueness[T any](slice LexicallyComparable[T]) ElementValidationFunc[T] {
	var prev *T
	var prevIndex int

	return func(index int, next T) error {
		if prev == nil {
			prev = &next
			prevIndex = index
		} else {
			switch slice.LexicalCompare(*prev, next) {
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
