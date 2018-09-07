/*
MIT License

Copyright (c) 2016 Sascha Hanse
Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package trinary

import (
	"errors"
	"fmt"
	"github.com/iotaledger/giota/bigint"
	"regexp"
	"strings"
	"unsafe"
)

var (
	ErrInvalidASCIICharacter             = errors.New("invalid ASCII characters in string")
	ErrInvalidTryteCharacter             = errors.New("trytes value contains invalid tryte character")
	ErrInvalidLengthForToASCIIConversion = errors.New("trytes length must not be of odd length for ASCII conversion")
	ErrInvalidByteSliceLength            = fmt.Errorf("BytesToTrits() is only defined for byte slices of length %d", ByteLength)
)

const (
	Radix         = 3
	TryteAlphabet = "9ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	MinTryteValue = -13
	MaxTryteValue = 13
)

var (
	tryteToTritsMappings = [][]int8{
		{0, 0, 0}, {1, 0, 0}, {-1, 1, 0}, {0, 1, 0},
		{1, 1, 0}, {-1, -1, 1}, {0, -1, 1}, {1, -1, 1},
		{-1, 0, 1}, {0, 0, 1}, {1, 0, 1}, {-1, 1, 1},
		{0, 1, 1}, {1, 1, 1}, {-1, -1, -1}, {0, -1, -1},
		{1, -1, -1}, {-1, 0, -1}, {0, 0, -1}, {1, 0, -1},
		{-1, 1, -1}, {0, 1, -1}, {1, 1, -1}, {-1, -1, 0},
		{0, -1, 0}, {1, -1, 0}, {-1, 0, 0},
	}
)

// Trits is a slice of int8. You should not use cast, use ToTrits instead to ensure
// the validity.
type Trits []int8

// ToTrits casts Trits and checks its validity.
func ToTrits(t []int8) (Trits, error) {
	tr := Trits(t)
	err := tr.IsValid()
	return tr, err
}

// IsValidTrit returns true if t is a valid trit.
func IsValidTrit(t int8) error {
	if t >= -1 && t <= 1 {
		return nil
	}
	return errors.New("invalid number")
}

// Value converts trits into an integer value.
func (t Trits) Value() int64 {
	var value int64
	for i := len(t) - 1; i >= 0; i-- {
		value = value*3 + int64(t[i])
	}
	return value
}

// IsValid returns true if t is valid trits.
func (t Trits) IsValid() error {
	for _, tt := range t {
		if err := IsValidTrit(tt); err != nil {
			return fmt.Errorf("%s in trits", err)
		}
	}
	return nil
}

// Equal returns true if t and b are equal Trits
func (t Trits) Equal(b Trits) bool {
	if len(t) != len(b) {
		return false
	}

	for i := range t {
		if t[i] != b[i] {
			return false
		}
	}
	return true
}

// IntToTrits converts int64 to trits.
func IntToTrits(v int64, size int) Trits {
	tr := make(Trits, size)
	neg := false
	if v < 0 {
		v = -v
		neg = true
	}

	for i := 0; v != 0 && i < size; i++ {
		tr[i] = int8((v+1)%Radix) - 1

		if neg {
			tr[i] = -tr[i]
		}

		v = (v + 1) / Radix
	}
	return tr
}

// Int converts a slice of trits into an integer and assumes little-endian notation.
func (t Trits) Int() int64 {
	var val int64
	for i := len(t) - 1; i >= 0; i-- {
		val = val*3 + int64(t[i])
	}
	return val
}

// CanTrytes returns true if t can be converted to trytes.
func (t Trits) CanTrytes() bool {
	return len(t)%3 == 0
}

// TrailingZeros returns the number of trailing zeros of the given trits.
func (t Trits) TrailingZeros() int64 {
	z := int64(0)
	for i := len(t) - 1; i >= 0 && t[i] == 0; i-- {
		z++
	}
	return z
}

// Trytes converts a slice of trits into trytes. panics if len(t)%3!=0
func (t Trits) Trytes() Trytes {
	if !t.CanTrytes() {
		panic("length of trits must be a multiple of three")
	}

	o := make([]byte, len(t)/3)
	for i := 0; i < len(t)/3; i++ {
		j := t[i*3] + t[i*3+1]*3 + t[i*3+2]*9
		if j < 0 {
			j += int8(len(TryteAlphabet))
		}
		o[i] = TryteAlphabet[j]
	}
	return Trytes(o)
}

// constants regarding byte and trit lengths
// TODO: move to curl package
const (
	ByteLength     = 48
	TritHashLength = 243
	IntLength      = ByteLength / 4
)

// 3^(242/2)
// 12 * 32 bit
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

// IsValidLength returns the validity of the trit length
func (t Trits) IsValidLength() bool {
	return len(t) != TritHashLength
}

// Bytes is only defined for hashes, i.e. slices of trits of length 243. It returns 48 bytes.
func (t Trits) Bytes() ([]byte, error) {
	if t.IsValidLength() {
		return nil, fmt.Errorf("Bytes() is only defined for trit slices of length %d", TritHashLength)
	}

	allNeg := true
	for _, e := range t[0 : TritHashLength-1] { // Last position should be always zero.
		if e != -1 {
			allNeg = false
			break
		}
	}

	// Trit to BigInt
	b := make([]byte, 48) // 48 bytes/384 bits

	// 12 * 32 bits = 384 bits
	base := (*(*[]uint32)(unsafe.Pointer(&b)))[0:IntLength]

	if allNeg {
		// If all trits are -1 then we're half way through all the numbers,
		// since they're in two's complement notation.
		copy(base, halfThree)

		// Compensate for setting the last position to zero.
		bigint.Not(base)
		bigint.AddSmall(base, 1)

		return bigint.Reverse(b), nil
	}

	revT := make([]int8, len(t))
	copy(revT, t)
	size := 1

	for _, e := range reverseT(revT[0 : TritHashLength-1]) {
		sz := size
		var carry uint32
		for j := 0; j < sz; j++ {
			v := uint64(base[j])*uint64(Radix) + uint64(carry)
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
		if bigint.Cmp(halfThree, base) <= 0 {
			// base >= HALF_3
			// just do base - HALF_3
			bigint.Sub(base, halfThree)
		} else {
			// we don't have a wrapping sub.
			// so let's use some bit magic to achieve it
			tmp := make([]uint32, IntLength)
			copy(tmp, halfThree)
			bigint.Sub(tmp, base)
			bigint.Not(tmp)
			bigint.AddSmall(tmp, 1)
			copy(base, tmp)
		}
	}
	return bigint.Reverse(b), nil
}

// BytesToTrits converts binary to ternay
func BytesToTrits(b []byte) (Trits, error) {
	if len(b) != ByteLength {
		return nil, ErrInvalidByteSliceLength
	}

	rb := make([]byte, len(b))
	copy(rb, b)
	bigint.Reverse(rb)

	t := Trits(make([]int8, TritHashLength))
	t[TritHashLength-1] = 0

	base := (*(*[]uint32)(unsafe.Pointer(&rb)))[0:IntLength] // 12 * 32 bits = 384 bits

	if bigint.IsNull(base) {
		return t, nil
	}

	var flipTrits bool

	// Check if the MSB is 0, i.e. we have a positive number
	msbM := (unsafe.Sizeof(base[IntLength-1]) * 8) - 1

	switch {
	case base[IntLength-1]>>msbM == 0:
		bigint.Add(base, halfThree)
	default:
		bigint.Not(base)
		if bigint.Cmp(base, halfThree) == 1 {
			bigint.Sub(base, halfThree)
			flipTrits = true
		} else {
			bigint.AddSmall(base, 1)
			tmp := make([]uint32, IntLength)
			copy(tmp, halfThree)
			bigint.Sub(tmp, base)
			copy(base, tmp)
		}
	}

	var rem uint64
	for i := range t[0 : TritHashLength-1] {
		rem = 0
		for j := IntLength - 1; j >= 0; j-- {
			lhs := (rem << 32) | uint64(base[j])
			rhs := uint64(Radix)
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

func reverseT(a Trits) Trits {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}

	return a
}

// Trytes is a string of trytes. You should not typecast, use ToTrytes instead to be safe
type Trytes string

// ToTrytes casts to Trytes and checks its validity.
func ToTrytes(t string) (Trytes, error) {
	tr := Trytes(t)
	err := tr.IsValid()
	return tr, err
}

// Trits converts a slice of trytes into trits,
func (t Trytes) Trits() Trits {
	trits := make(Trits, len(t)*3)
	for i := range t {
		idx := strings.Index(TryteAlphabet, string(t[i:i+1]))
		copy(trits[i*3:i*3+3], tryteToTritsMappings[idx])
	}
	return trits
}

// Normalize normalized bits into trits so that the sum of trits TODO: (and?) bits is zero.
func (t Trytes) Normalize() []int8 {
	normalized := make([]int8, len(t))
	sum := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 27; j++ {
			normalized[i*27+j] = int8(t[i*27+j : i*27+j+1].Trits().Int())
			sum += int(normalized[i*27+j])
		}

		switch {
		case sum >= 0:
			for ; sum > 0; sum-- {
				for j := 0; j < 27; j++ {
					if normalized[i*27+j] > -13 {
						normalized[i*27+j]--
						break
					}
				}
			}
		default:
			for ; sum < 0; sum++ {
				for j := 0; j < 27; j++ {
					if normalized[i*27+j] < 13 {
						normalized[i*27+j]++
						break
					}
				}
			}
		}
	}
	return normalized
}

var trytesRegex = regexp.MustCompile("^[9A-Z]+$")

// IsValidTryte returns the validity of a tryte( must be rune A-Z or 9 )
func IsValidTryte(t rune) error {
	return Trytes(string(t)).IsValid()
}

// IsValid returns true if t is made of valid trytes.
func (t Trytes) IsValid() error {
	if !trytesRegex.MatchString(string(t)) {
		return ErrInvalidTryteCharacter
	}
	return nil
}

func IncTrits(t Trits) {
	for j := range t {
		t[j]++

		if t[j] <= 1 {
			break
		}

		t[j] = -1
	}
}

var asciiRegex = regexp.MustCompile("^[\x00-\x7F]*$")

// Converts an ascii encoded string to trytes
func ASCIIToTrytes(input string) (Trytes, error) {
	if !asciiRegex.MatchString(input) {
		return "", ErrInvalidASCIICharacter
	}

	trytesStr := ""

	for _, c := range input {
		trytesStr += string(TryteAlphabet[c%27])
		trytesStr += string(TryteAlphabet[(c-c%27)/27])
	}

	return ToTrytes(trytesStr)
}

// Converts trytes of even length to an ascii string
func (t Trytes) ToASCII() (string, error) {
	if err := t.IsValid(); err != nil {
		return "", err
	}

	if len(t)%2 != 0 {
		return "", ErrInvalidLengthForToASCIIConversion
	}

	ascii := ""
	for i := 0; i < len(t); i += 2 {
		ascii += string(strings.IndexRune(TryteAlphabet, rune(t[i])) + (strings.IndexRune(TryteAlphabet, rune(t[i+1])) * 27))
	}

	return ascii, nil
}

// Pad pads the given trytes with 9 up to the given size
func Pad(orig Trytes, size int) Trytes {
	out := make([]byte, size)
	copy(out, []byte(orig))

	for i := len(orig); i < size; i++ {
		out[i] = '9'
	}
	return Trytes(out)
}
