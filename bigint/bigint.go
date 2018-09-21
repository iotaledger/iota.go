package bigint

import (
	"math"
)

func Add(b []uint32, rh []uint32) {
	if len(b) != len(rh) {
		panic("not defined for differently sized slices")
	}

	carry := false

	for i := range b {
		v, c := FullAdd(b[i], rh[i], carry)
		b[i] = uint32(v)
		carry = c
	}
}

func Sub(b []uint32, rh []uint32) {
	if len(b) != len(rh) {
		panic("not defined for differently sized slices")
	}

	noborrow := true

	for i := range b {
		v, c := FullAdd(b[i], ^rh[i], noborrow)
		b[i] = uint32(v)
		noborrow = c
	}

	if !noborrow {
		panic("could not subtract without leftovers")
	}
}

func Not(b []uint32) {
	for i := range b {
		b[i] = ^b[i]
	}
}

func IsNull(b []uint32) bool {
	for i := range b {
		if b[i] != 0 {
			return false
		}
	}
	return true
}

func Cmp(lh, rh []uint32) int {
	if len(lh) != len(rh) {
		panic("not defined for differently sized slices")
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

// bigIntAddSmall adds a small number to a big int and returns the index
// of the last carry over.
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

func FullAdd(lh, rh uint32, carry bool) (uint, bool) {
	v, c1 := AddWithOverflow(lh, rh)
	var c2 bool
	if carry {
		v, c2 = AddWithOverflow(uint32(v), 1)
	}

	return v, c1 || c2
}

func AddWithOverflow(lh, rh uint32) (uint, bool) {
	return uint(lh + rh), lh > math.MaxUint32-rh
}

func Reverse(a []byte) []byte {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}

	return a
}

func ReverseU(a []uint32) []uint32 {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}

	return a
}
