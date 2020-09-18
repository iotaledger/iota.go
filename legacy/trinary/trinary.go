// Package trinary provides functions for validating and converting Trits and Trytes.
package trinary

import (
	"bytes"
	"math"
	"strings"

	"github.com/iotaledger/iota.go/legacy"
	iotaGoMath "github.com/iotaledger/iota.go/math"
	"github.com/pkg/errors"
)

var (
	// TryteValueToTritsLUT is a lookup table to convert tryte values into trits.
	TryteValueToTritsLUT = [legacy.TryteRadix][legacy.TritsPerTryte]int8{
		{-1, -1, -1}, {0, -1, -1}, {1, -1, -1}, {-1, 0, -1}, {0, 0, -1}, {1, 0, -1},
		{-1, 1, -1}, {0, 1, -1}, {1, 1, -1}, {-1, -1, 0}, {0, -1, 0}, {1, -1, 0},
		{-1, 0, 0}, {0, 0, 0}, {1, 0, 0}, {-1, 1, 0}, {0, 1, 0}, {1, 1, 0},
		{-1, -1, 1}, {0, -1, 1}, {1, -1, 1}, {-1, 0, 1}, {0, 0, 1}, {1, 0, 1},
		{-1, 1, 1}, {0, 1, 1}, {1, 1, 1},
	}

	// TryteValueToTyteLUT is a lookup table to convert tryte values into trytes.
	TryteValueToTyteLUT = [legacy.TryteRadix]byte{'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M'}

	// TryteToTryteValueLUT is a lookup table to convert trytes into tryte values.
	TryteToTryteValueLUT = [...]int8{
		0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		-13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1,
	}

	// Pow27LUT is a Look-up-table for Decoding Trits to int64
	Pow27LUT = []int64{1,
		27,
		729,
		19683,
		531441,
		14348907,
		387420489,
		10460353203,
		282429536481,
		7625597484987,
		205891132094649,
		5559060566555523,
		150094635296999136,
		4052555153018976256}

	encodedZero = []int8{1, 0, 0, -1}
)

// Trits is a slice of int8. You should not use cast, use NewTrits instead to ensure the validity.
type Trits = []int8

// Trytes is a string of trytes. Use NewTrytes() instead of typecasting.
type Trytes = string

// Hash represents a trinary hash
type Hash = Trytes

// Hashes is a slice of Hash.
type Hashes = []Hash

// ValidTrit returns true if t is a valid trit.
func ValidTrit(t int8) bool {
	return t >= -1 && t <= 1
}

// ValidTrits returns true if t is valid trits (non-empty and -1, 0 or 1).
func ValidTrits(trits Trits) error {
	if len(trits) == 0 {
		return errors.Wrap(legacy.ErrInvalidTrit, "trits slice is empty")
	}
	for i, trit := range trits {
		if !ValidTrit(trit) {
			return errors.Wrapf(legacy.ErrInvalidTrit, "at index %d", i)
		}
	}
	return nil
}

// NewTrits casts Trits and checks its validity.
func NewTrits(t []int8) (Trits, error) {
	err := ValidTrits(t)
	return t, err
}

// TritsEqual returns true if t and b are equal Trits.
func TritsEqual(a Trits, b Trits) (bool, error) {
	if err := ValidTrits(a); err != nil {
		return false, err
	}
	if err := ValidTrits(b); err != nil {
		return false, err
	}

	if len(a) != len(b) {
		return false, nil
	}

	for i := range a {
		if a[i] != b[i] {
			return false, nil
		}
	}
	return true, nil
}

// ReverseTrits reverses the given trits.
func ReverseTrits(trits Trits) Trits {
	for left, right := 0, len(trits)-1; left < right; left, right = left+1, right-1 {
		trits[left], trits[right] = trits[right], trits[left]
	}

	return trits
}

// TrailingZeros returns the number of trailing zeros of the given trits.
func TrailingZeros(trits Trits) int {
	var z int
	for i := len(trits) - 1; i >= 0 && trits[i] == 0; i-- {
		z++
	}
	return z
}

// roundUpToTryteMultiple rounds the given number up the the nearest multiple of 3 to make a valid tryte count.
func roundUpToTryteMultiple(n uint) uint {
	rem := n % legacy.TritsPerTryte
	if rem == 0 {
		return n
	}
	return n + legacy.TritsPerTryte - rem
}

// MinTrits returns the length of trits needed to encode the value.
func MinTrits(value int64) int {
	valueAbs := iotaGoMath.AbsInt64(value)

	var vp uint64
	var num int
	switch {
	case valueAbs >= 308836698141973:
		vp = 308836698141973
		num = 31
	case valueAbs >= 5230176601:
		vp = 5230176601
		num = 21
	case valueAbs >= 88573:
		vp = 88573
		num = 11
	default:
		vp = 1
		num = 1
	}

	for valueAbs > vp {
		vp = vp*legacy.TrinaryRadix + 1
		num++
	}
	return num
}

// IntToTrits converts int64 to a slice of trits.
func IntToTrits(value int64) Trits {
	numTrits := MinTrits(value)
	numTrytes := (numTrits + legacy.TritsPerTryte - 1) / legacy.TritsPerTryte
	trits := MustTrytesToTrits(IntToTrytes(value, numTrytes))
	return trits[:numTrits]
}

// TritsToInt converts a slice of trits into an integer and assumes little-endian notation.
func TritsToInt(t Trits) int64 {
	var val int64
	for i := len(t) - 1; i >= 0; i-- {
		val = val*legacy.TrinaryRadix + int64(t[i])
	}
	return val
}

// IntToTrytes converts int64 to a slice of trytes.
func IntToTrytes(value int64, trytesCnt int) Trytes {
	negative := value < 0
	v := iotaGoMath.AbsInt64(value)

	var trytes strings.Builder
	trytes.Grow(trytesCnt)

	for i := 0; i < trytesCnt; i++ {
		if v == 0 {
			trytes.WriteByte('9')
			continue
		}

		v += legacy.TryteRadix / 2
		tryte := int8(v%legacy.TryteRadix) - legacy.TryteRadix/2
		v /= legacy.TryteRadix
		if negative {
			tryte = -tryte
		}
		trytes.WriteByte(MustTryteValueToTryte(tryte))
	}
	return trytes.String()
}

// TrytesToInt converts a slice of trytes to int64.
func TrytesToInt(t Trytes) int64 {
	// ignore tailing 9s
	var i int
	for i = len(t) - 1; i >= 0; i-- {
		if t[i] != '9' {
			break
		}
	}

	var val int64
	for ; i >= 0; i-- {
		val = val*legacy.TryteRadix + int64(MustTryteToTryteValue(t[i]))
	}
	return val
}

// CanTritsToTrytes returns true if t can be converted to trytes.
func CanTritsToTrytes(trits Trits) bool {
	if len(trits) == 0 {
		return false
	}
	return len(trits)%legacy.TritsPerTryte == 0
}

// MustTryteValueToTryte converts the value of a tryte v in [-13,13] to a tryte char in [9A-Z].
// It panics when v is an invalid value.
func MustTryteValueToTryte(v int8) byte {
	idx := uint(v - legacy.MinTryteValue)
	if idx >= uint(len(TryteValueToTyteLUT)) {
		panic(legacy.ErrInvalidTrytes)
	}
	return TryteValueToTyteLUT[idx]
}

// MustTryteToTryteValue converts a tryte char t in [9A-Z] to a tryte value in [-13,13].
// It panics when t is an invalid tryte.
func MustTryteToTryteValue(t byte) int8 {
	idx := uint(t - '9')
	if idx >= uint(len(TryteToTryteValueLUT)) {
		panic(legacy.ErrInvalidTrytes)
	}
	return TryteToTryteValueLUT[idx]
}

// TritsToTrytes converts a slice of trits into trytes. Returns an error if len(t)%3!=0
func TritsToTrytes(trits Trits) (Trytes, error) {
	if err := ValidTrits(trits); err != nil {
		return "", err
	}
	if !CanTritsToTrytes(trits) {
		return "", errors.Wrap(legacy.ErrInvalidTritsLength, "trits slice size must be a multiple of 3")
	}
	return MustTritsToTrytes(trits), nil
}

// MustTritsToTrytes converts a slice of trits into trytes.
// Performs no validation on the input trits and might therefore return an invalid trytes representation
// (without a panic).
func MustTritsToTrytes(trits Trits) Trytes {
	trytes := make([]byte, len(trits)/legacy.TritsPerTryte)
	for i := range trytes {
		v := MustTritsToTryteValue(trits[i*legacy.TritsPerTryte:])
		trytes[i] = MustTryteValueToTryte(v)
	}
	return string(trytes)
}

func validTryte(t rune) bool {
	return (t >= 'A' && t <= 'Z') || t == '9'
}

// ValidTryte returns the validity of a tryte (must be rune A-Z or 9)
func ValidTryte(t rune) error {
	if !validTryte(t) {
		return legacy.ErrInvalidTrytes
	}
	return nil
}

// ValidTrytes returns true if t is made of valid trytes.
func ValidTrytes(trytes Trytes) error {
	if trytes == "" {
		return legacy.ErrInvalidTrytes
	}
	for _, tryte := range trytes {
		if !validTryte(tryte) {
			return legacy.ErrInvalidTrytes
		}
	}
	return nil
}

// NewTrytes casts to Trytes and checks its validity.
func NewTrytes(s string) (Trytes, error) {
	err := ValidTrytes(s)
	return s, err
}

// TrytesToTrits converts a slice of trytes into trits.
func TrytesToTrits(trytes Trytes) (Trits, error) {
	if err := ValidTrytes(trytes); err != nil {
		return nil, err
	}
	return MustTrytesToTrits(trytes), nil
}

// MustTrytesToTrits converts a slice of trytes into trits.
// Performs no validation on the provided inputs (therefore might return an invalid representation) and might panic.
func MustTrytesToTrits(trytes Trytes) Trits {
	trits := make(Trits, len(trytes)*legacy.TritsPerTryte)
	for i := 0; i < len(trytes); i++ {
		MustPutTryteTrits(trits[i*legacy.TritsPerTryte:], MustTryteToTryteValue(trytes[i]))
	}
	return trits
}

// MustTritsToTryteValue converts a slice of 3 into its corresponding value.
// It performs no validation on the provided inputs (therefore might return an invalid representation) and might panic.
func MustTritsToTryteValue(trits Trits) int8 {
	_ = trits[2] // bounds check hint to compiler
	return trits[0] + trits[1]*3 + trits[2]*9
}

// MustPutTryteTrits converts v in [-13,13] to its corresponding 3-trit value and writes this to trits.
// It panics on invalid input.
func MustPutTryteTrits(trits []int8, v int8) {
	idx := v - legacy.MinTryteValue
	_ = trits[2] // early bounds check to guarantee safety of writes below
	trits[0] = TryteValueToTritsLUT[idx][0]
	trits[1] = TryteValueToTritsLUT[idx][1]
	trits[2] = TryteValueToTritsLUT[idx][2]
}

// CanBeHash returns the validity of the trit length.
func CanBeHash(trits Trits) bool {
	return len(trits) == legacy.HashTrinarySize
}

// Pad pads the given trytes with 9s up to the given size.
func Pad(trytes Trytes, n int) (Trytes, error) {
	if len(trytes) > 0 {
		if err := ValidTrytes(trytes); err != nil {
			return "", err
		}
	}
	return MustPad(trytes, n), nil
}

// MustPad pads the given trytes with 9s up to the given size.
// Performs no validation on the provided inputs (therefore might return an invalid representation) and might panic.
func MustPad(trytes Trytes, n int) Trytes {
	if len(trytes) >= n {
		return trytes
	}

	var result strings.Builder
	result.Grow(n)

	result.WriteString(trytes)
	result.Write(bytes.Repeat([]byte{'9'}, n-len(trytes)))

	return result.String()
}

// PadTrits pads the given trits with 0 up to the given size.
func PadTrits(trits Trits, n int) (Trits, error) {
	if len(trits) > 0 {
		if err := ValidTrits(trits); err != nil {
			return nil, err
		}
	}
	return MustPadTrits(trits, n), nil
}

// MustPadTrits pads the given trits with 0 up to the given size.
// Performs no validation on the provided inputs (therefore might return an invalid representation) and might panic.
func MustPadTrits(trits Trits, n int) Trits {
	if len(trits) >= n {
		return trits
	}

	result := make(Trits, n)
	copy(result, trits)
	return result
}

// Sum returns the sum of two trits.
func Sum(a int8, b int8) int8 {
	s := a + b

	switch s {
	case 2:
		return -1
	case -2:
		return 1
	default:
		return s
	}
}

func cons(a int8, b int8) int8 {
	if a == b {
		return a
	}

	return 0
}

func any(a int8, b int8) int8 {
	s := a + b

	if s > 0 {
		return 1
	}

	if s < 0 {
		return -1
	}

	return 0
}

func fullAdd(a int8, b int8, c int8) [2]int8 {
	sA := Sum(a, b)
	cA := cons(a, b)
	cB := cons(sA, c)
	cOut := any(cA, cB)
	sOut := Sum(sA, c)
	return [2]int8{sOut, cOut}
}

// AddTrits adds a to b.
func AddTrits(a Trits, b Trits) Trits {
	maxLen := int64(math.Max(float64(len(a)), float64(len(b))))
	if maxLen == 0 {
		return Trits{0}
	}
	out := make(Trits, maxLen)
	var aI, bI, carry int8

	for i := 0; i < len(out); i++ {
		if i < len(a) {
			aI = a[i]
		} else {
			aI = 0
		}
		if i < len(b) {
			bI = b[i]
		} else {
			bI = 0
		}

		fA := fullAdd(aI, bI, carry)
		out[i] = fA[0]
		carry = fA[1]
	}
	return out
}
