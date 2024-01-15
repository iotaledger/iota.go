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
		switch {
		case prev == nil:
			prev = &next
			prevIndex = index
		// TODO: Optimize to return different error when lexical order vs uniquness is violated.
		case slice.LexicalCompare(*prev, next) > 0:
			return ierrors.Wrapf(ErrArrayValidationOrderViolatesLexicalOrder, "element %d should have been before element %d", index, prevIndex)
		case slice.LexicalCompare(*prev, next) == 0:
			// TODO: Error message.
			return ierrors.Wrapf(ErrArrayValidationViolatesUniqueness, "TODO: element %d should have been before element %d", index, prevIndex)
		default:
			prev = &next
			prevIndex = index
		}

		return nil
	}
}
