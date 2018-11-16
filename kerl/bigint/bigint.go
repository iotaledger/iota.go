// Package bigint implements a set of functions for big integer arithmetic.
package bigint

import (
	"github.com/pkg/errors"
	"math"
)

// Errors for bigint package.
var (
	ErrUnequallySizedSlices     = errors.New("operation not defined for differently sized slices")
	ErrSubtractionWithLeftovers = errors.New("could not subtract without leftovers")
)

// MustAdd adds the given big ints together.
func MustAdd(b []uint32, rh []uint32) {
	if len(b) != len(rh) {
		panic(ErrUnequallySizedSlices)
	}

	carry := false

	for i := range b {
		v, c := FullAdd(b[i], rh[i], carry)
		b[i] = uint32(v)
		carry = c
	}
}

// MustSub subtracts rh from b.
func MustSub(b []uint32, rh []uint32) {
	if len(b) != len(rh) {
		panic(ErrUnequallySizedSlices)
	}

	noborrow := true

	for i := range b {
		v, c := FullAdd(b[i], ^rh[i], noborrow)
		b[i] = uint32(v)
		noborrow = c
	}

	if !noborrow {
		panic(ErrSubtractionWithLeftovers)
	}
}

// Not negates the given big int value.
func Not(b []uint32) {
	for i := range b {
		b[i] = ^b[i]
	}
}

// IsNull checks whether the given big int value is null.
func IsNull(b []uint32) bool {
	for i := range b {
		if b[i] != 0 {
			return false
		}
	}
	return true
}

// MustCmp compares the given big ints with each other.
func MustCmp(lh, rh []uint32) int {
	if len(lh) != len(rh) {
		panic(ErrUnequallySizedSlices)
	}

	// put LSB first
	rlh := make([]uint32, len(lh))
	copy(rlh, lh)
	ReverseU(rlh)

	rrh := make([]uint32, len(rh))
	copy(rrh, rh)
	ReverseU(rrh)

	for i := range rlh {
		switch {
		case rlh[i] < rrh[i]:
			return -1
		case rlh[i] > rrh[i]:
			return 1
		}
	}
	return 0
}

// AddSmall adds a small number to a big int and returns the index of the last carry over.
func AddSmall(b []uint32, a uint32) int {
	v, carry := FullAdd(b[0], a, false)
	b[0] = uint32(v) // uint is at least 32 bit

	var i int
	for i = 1; carry; i++ {
		vi, c := FullAdd(b[i], 0, carry)
		b[i] = uint32(vi)
		carry = c
	}

	return i
}

// FullAdd adds left and right together and whether it overflowed.
func FullAdd(lh, rh uint32, carry bool) (uint, bool) {
	v, c1 := AddWithOverflow(lh, rh)
	var c2 bool
	if carry {
		v, c2 = AddWithOverflow(uint32(v), 1)
	}

	return v, c1 || c2
}

// AddWithOverflow returns left hand + right hand and whether it overflowed.
func AddWithOverflow(lh, rh uint32) (uint, bool) {
	return uint(lh + rh), lh > math.MaxUint32-rh
}

// Reverse reverses the given byte slice.
func Reverse(a []byte) []byte {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}

	return a
}

// ReverseU reverses the given uint32 slice.
func ReverseU(a []uint32) []uint32 {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}

	return a
}
