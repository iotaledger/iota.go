// Package bigint implements a set of functions for big integer arithmetic.
package bigint

import (
	"github.com/pkg/errors"
)

// Errors for bigint package.
var (
	ErrUnequallySizedSlices = errors.New("operation not defined for differently sized slices")
)

// MustAdd adds the given big ints together.
func MustAdd(b []uint32, rh []uint32) {
	if len(b) != len(rh) {
		panic(ErrUnequallySizedSlices)
	}

	carry := false
	for i := range b {
		b[i], carry = FullAdd(b[i], rh[i], carry)
	}
}

// MustSub subtracts rh from b.
func MustSub(b []uint32, rh []uint32) {
	if len(b) != len(rh) {
		panic(ErrUnequallySizedSlices)
	}

	noborrow := true
	for i := range b {
		b[i], noborrow = FullAdd(b[i], ^rh[i], noborrow)
	}
}

// IsNegative checks whether the given big int represents a negative number in two's complement.
func IsNegative(b []uint32) bool {
	ms := b[len(b)-1]
	return (ms >> 31) != 0
}

// MustCmp compares the given big ints with each other.
func MustCmp(lh, rh []uint32) int {
	if len(lh) != len(rh) {
		panic(ErrUnequallySizedSlices)
	}

	for i := len(lh) - 1; i >= 0; i-- {
		switch {
		case lh[i] < rh[i]:
			return -1
		case lh[i] > rh[i]:
			return 1
		}
	}
	return 0
}

// AddSmall adds a uint32 to a big int and returns the highest index that was changed.
func AddSmall(b []uint32, a uint32) int {
	v, carry := FullAdd(b[0], a, false)
	b[0] = v
	if !carry {
		return 0
	}
	for i := 1; i < len(b); i++ {
		b[i], carry = FullAdd(b[i], 0, carry)
		if !carry {
			return i
		}
	}
	return len(b)
}

// FullAdd returns the sum and whether the operation overflowed.
func FullAdd(lh, rh uint32, carry bool) (uint32, bool) {
	v, c1 := addCarry(lh, rh)
	var c2 bool
	if carry {
		v, c2 = addCarry(v, 1)
	}
	return v, c1 || c2
}

func addCarry(lh, rh uint32) (uint32, bool) {
	sum := lh + rh
	return sum, sum < lh
}
