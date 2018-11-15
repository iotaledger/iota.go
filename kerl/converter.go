// Package kerl implements the Kerl hashing function.
package kerl

import (
	"unsafe"

	"github.com/pkg/errors"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl/bigint"
	. "github.com/iotaledger/iota.go/trinary"
)

// 12 * 32 bit
// hex representation of (3^242)/2
var halfThree = []uint32{
	0xa5ce8964,
	0x9f007669,
	0x1484504f,
	0x3ade00d9,
	0x0c24486e,
	0x50979d57,
	0x79a4c702,
	0x48bbae36,
	0xa9f6808b,
	0xaa06a805,
	0xa87fabdf,
	0x5e69ebef,
}

// KerlTritsToBytes is only defined for hashes, i.e. slices of trits of length 243. It returns 48 bytes.
func KerlTritsToBytes(trits Trits) ([]byte, error) {
	if !CanBeHash(trits) {
		return nil, errors.Wrapf(ErrInvalidTritsLength, "must be %d in size", HashTrinarySize)
	}

	allNeg := true
	// last position should be always zero.
	for _, e := range trits[0 : HashTrinarySize-1] {
		if e != -1 {
			allNeg = false
			break
		}
	}

	// trit to BigInt
	b := make([]byte, 48) // 48 bytes/384 bits

	// 12 * 32 bits = 384 bits
	base := (*(*[]uint32)(unsafe.Pointer(&b)))[0:IntLength]

	if allNeg {
		// if all trits are -1 then we're half way through all the numbers,
		// since they're in two's complement notation.
		copy(base, halfThree)

		// compensate for setting the last position to zero.
		bigint.Not(base)
		bigint.AddSmall(base, 1)

		return bigint.Reverse(b), nil
	}

	revT := make([]int8, len(trits))
	copy(revT, trits)
	size := 1

	for _, e := range ReverseTrits(revT[0 : HashTrinarySize-1]) {
		sz := size
		var carry uint32
		for j := 0; j < sz; j++ {
			v := uint64(base[j])*uint64(TrinaryRadix) + uint64(carry)
			carry = uint32(v >> 32)
			base[j] = uint32(v)
		}

		if carry > 0 {
			base[sz] = carry
			size = size + 1
		}

		trit := uint32(e + 1)

		ns := bigint.AddSmall(base, trit)
		if ns > size {
			size = ns
		}
	}

	if !bigint.IsNull(base) {
		if bigint.MustCmp(halfThree, base) <= 0 {
			// base >= HALF_3
			// just do base - HALF_3
			bigint.MustSub(base, halfThree)
		} else {
			// we don'trits have a wrapping sub.
			// so let's use some bit magic to achieve it
			tmp := make([]uint32, IntLength)
			copy(tmp, halfThree)
			bigint.MustSub(tmp, base)
			bigint.Not(tmp)
			bigint.AddSmall(tmp, 1)
			copy(base, tmp)
		}
	}
	return bigint.Reverse(b), nil
}

// KerlBytesToTrits converts binary to trinary
func KerlBytesToTrits(b []byte) (Trits, error) {
	if len(b) != HashBytesSize {
		return nil, errors.Wrapf(ErrInvalidBytesLength, "must be %d in size", HashBytesSize)
	}

	rb := make([]byte, len(b))
	copy(rb, b)
	bigint.Reverse(rb)

	t := Trits(make([]int8, HashTrinarySize))
	t[HashTrinarySize-1] = 0

	base := (*(*[]uint32)(unsafe.Pointer(&rb)))[0:IntLength] // 12 * 32 bits = 384 bits

	if bigint.IsNull(base) {
		return t, nil
	}

	var flipTrits bool

	// Check if the MSB is 0, i.e. we have a positive number
	msbM := (unsafe.Sizeof(base[IntLength-1]) * 8) - 1

	switch {
	case base[IntLength-1]>>msbM == 0:
		bigint.MustAdd(base, halfThree)
	default:
		bigint.Not(base)
		if bigint.MustCmp(base, halfThree) == 1 {
			bigint.MustSub(base, halfThree)
			flipTrits = true
		} else {
			bigint.AddSmall(base, 1)
			tmp := make([]uint32, IntLength)
			copy(tmp, halfThree)
			bigint.MustSub(tmp, base)
			copy(base, tmp)
		}
	}

	var rem uint64
	for i := range t[0 : HashTrinarySize-1] {
		rem = 0
		for j := IntLength - 1; j >= 0; j-- {
			lhs := (rem << 32) | uint64(base[j])
			rhs := uint64(TrinaryRadix)
			q := uint32(lhs / rhs)
			r := uint32(lhs % rhs)
			base[j] = q
			rem = uint64(r)
		}
		t[i] = int8(rem) - 1
	}

	if flipTrits {
		for i := range t {
			t[i] = -t[i]
		}
	}

	return t, nil
}
