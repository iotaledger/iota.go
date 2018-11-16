// Package trinary provides functions for validating and converting Trits and Trytes.
package trinary

import (
	"math"
	"regexp"
	"strings"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/pkg/errors"
)

var (
	// TryteToTritsLUT is a Look-up-table for Trytes to Trits conversion.
	TryteToTritsLUT = [][]int8{
		{0, 0, 0}, {1, 0, 0}, {-1, 1, 0}, {0, 1, 0},
		{1, 1, 0}, {-1, -1, 1}, {0, -1, 1}, {1, -1, 1},
		{-1, 0, 1}, {0, 0, 1}, {1, 0, 1}, {-1, 1, 1},
		{0, 1, 1}, {1, 1, 1}, {-1, -1, -1}, {0, -1, -1},
		{1, -1, -1}, {-1, 0, -1}, {0, 0, -1}, {1, 0, -1},
		{-1, 1, -1}, {0, 1, -1}, {1, 1, -1}, {-1, -1, 0},
		{0, -1, 0}, {1, -1, 0}, {-1, 0, 0},
	}

	byteRadix = [5]int8{1, 3, 9, 27, 81}
)

// Trits is a slice of int8. You should not use cast, use NewTrits instead to ensure the validity.
type Trits = []int8

// ValidTrit returns true if t is a valid trit.
func ValidTrit(t int8) bool {
	if t == -1 || t == 0 || t == 1 {
		return true
	}
	return false
}

// ValidTrits returns true if t is valid trits.
func ValidTrits(t Trits) error {
	for i, tt := range t {
		if valid := ValidTrit(tt); !valid {
			return errors.Wrapf(ErrInvalidTrit, "at index %d", i)
		}
	}
	return nil
}

// NewTrits casts Trits and checks its validity.
func NewTrits(t []int8) (Trits, error) {
	err := ValidTrits(t)
	return t, err
}

// TritsEqual returns true if t and b are equal Trits
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

// IntToTrits converts int64 to trits.
func IntToTrits(value int64) Trits {
	if value == 0 {
		return Trits{0}
	}
	var dest Trits
	if value != 0 {
		dest = make(Trits, int(1+math.Floor(math.Log(2*math.Max(1, math.Abs(float64(value))))/math.Log(3))))
	} else {
		dest = make(Trits, 0)
	}

	var absoluteValue int64
	if value < 0 {
		absoluteValue = -value
	} else {
		absoluteValue = value
	}

	i := 0
	for absoluteValue > 0 {
		remainder := absoluteValue % TrinaryRadix
		absoluteValue = int64(math.Floor(float64(absoluteValue / TrinaryRadix)))

		if remainder > MaxTritValue {
			remainder = MinTritValue
			absoluteValue++
		}

		dest[i] = int8(remainder)
		i++
	}

	if value < 0 {
		for i := 0; i < len(dest); i++ {
			dest[i] = -dest[i]
		}
	}

	return dest
}

// TritsToInt converts a slice of trits into an integer and assumes little-endian notation.
func TritsToInt(t Trits) int64 {
	var val int64
	for i := len(t) - 1; i >= 0; i-- {
		val = val*3 + int64(t[i])
	}
	return val
}

// CanTritsToTrytes returns true if t can be converted to trytes.
func CanTritsToTrytes(trits Trits) bool {
	if len(trits) == 0 {
		return false
	}
	return len(trits)%3 == 0
}

// TrailingZeros returns the number of trailing zeros of the given trits.
func TrailingZeros(trits Trits) int64 {
	z := int64(0)
	for i := len(trits) - 1; i >= 0 && trits[i] == 0; i-- {
		z++
	}
	return z
}

// TritsToTrytes converts a slice of trits into trytes. Returns an error if len(t)%3!=0
func TritsToTrytes(trits Trits) (Trytes, error) {
	if !CanTritsToTrytes(trits) {
		return "", errors.Wrap(ErrInvalidTritsLength, "trits slice size must be a multiple of 3")
	}

	o := make([]byte, len(trits)/3)
	for i := 0; i < len(trits)/3; i++ {
		j := trits[i*3] + trits[i*3+1]*3 + trits[i*3+2]*9
		if j < 0 {
			j += int8(len(TryteAlphabet))
		}
		o[i] = TryteAlphabet[j]
	}
	return Trytes(o), nil
}

// MustTritsToTrytes converts a slice of trits into trytes. Panics if len(t)%3!=0
func MustTritsToTrytes(trits Trits) Trytes {
	trytes, err := TritsToTrytes(trits)
	if err != nil {
		panic(err)
	}
	return trytes
}

// CanBeHash returns the validity of the trit length.
func CanBeHash(trits Trits) bool {
	return len(trits) == HashTrinarySize
}

// TrytesToBytes is only defined for hashes (81 Trytes). It returns 48 bytes.
func TrytesToBytes(trytes Trytes) ([]byte, error) {
	trits, err := TrytesToTrits(trytes)
	if err != nil {
		return nil, err
	}
	return TritsToBytes(trits), nil
}

// BytesToTrytes converts bytes to Trytes. Returns an error if the bytes slice is not 48 in length.
func BytesToTrytes(bytes []byte) (Trytes, error) {
	trits, err := BytesToTrits(bytes)
	if err != nil {
		return "", err
	}
	return TritsToTrytes(trits)
}

// TritsToBytes packs an array of trits into an array of bytes (5 packed trits in 1 byte)
func TritsToBytes(trits Trits) (bytes []byte) {
	tritsLength := len(trits)
	bytesLength := (tritsLength + NumberOfTritsInAByte - 1) / NumberOfTritsInAByte

	bytes = make([]byte, bytesLength)

	tritIdx := bytesLength * NumberOfTritsInAByte
	for byteNum := bytesLength - 1; byteNum >= 0; byteNum-- {
		var value int8 = 0

		for i := 0; i < NumberOfTritsInAByte; i++ {
			tritIdx--

			if tritIdx < tritsLength {
				value = value*Radix + trits[tritIdx]
			}
		}
		bytes[byteNum] = byte(value)
	}
	return bytes
}

// BytesToTrits unpacks an array of bytes into an array of trits
func BytesToTrits(bytes []byte, numTrits ...int) (trits Trits, err error) {
	bytesLength := len(bytes)
	tritsLength := bytesLength * NumberOfTritsInAByte

	if len(numTrits) > 0 {
		tritsLength = numTrits[0]

		minTritLength := (bytesLength-1)*NumberOfTritsInAByte + 1
		maxTritLength := bytesLength * NumberOfTritsInAByte
		if tritsLength < minTritLength || tritsLength > maxTritLength {
			return nil, errors.Wrapf(ErrInvalidTritsLength, "must be %d-%d in size", minTritLength, maxTritLength)
		}
	}

	trits = make(Trits, tritsLength)

	for byteNum := 0; byteNum < bytesLength; byteNum++ {
		value := int8(bytes[byteNum])

		tritOffset := byteNum * NumberOfTritsInAByte

		for tritNum := NumberOfTritsInAByte - 1; tritNum >= 0; tritNum-- {
			var trit int8 = 0

			tritIdx := tritOffset + tritNum

			if tritIdx < tritsLength {
				byteRadixHalf := byteRadix[tritNum] >> 1
				if value > byteRadixHalf {
					value -= byteRadix[tritNum]
					trit = 1
				} else if value < (-byteRadixHalf) {
					value += byteRadix[tritNum]
					trit = -1
				}

				trits[tritIdx] = trit
			}
		}
	}
	return trits, nil
}

// ReverseTrits reverses the given trits.
func ReverseTrits(trits Trits) Trits {
	for left, right := 0, len(trits)-1; left < right; left, right = left+1, right-1 {
		trits[left], trits[right] = trits[right], trits[left]
	}

	return trits
}

// Trytes is a string of trytes. Use NewTrytes() instead of typecasting.
type Trytes = string

// Hash represents a trinary hash
type Hash = Trytes

// Hashes is a slice of Hash.
type Hashes = []Hash

var trytesRegex = regexp.MustCompile("^[9A-Z]+$")

// ValidTrytes returns true if t is made of valid trytes.
func ValidTrytes(trytes Trytes) error {
	if !trytesRegex.MatchString(string(trytes)) {
		return ErrInvalidTrytes
	}
	return nil
}

// ValidTryte returns the validity of a tryte (must be rune A-Z or 9)
func ValidTryte(t rune) error {
	return ValidTrytes(string(t))
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
	trits := make(Trits, len(trytes)*3)
	for i := range trytes {
		idx := strings.Index(TryteAlphabet, string(trytes[i:i+1]))
		copy(trits[i*3:i*3+3], TryteToTritsLUT[idx])
	}
	return trits, nil
}

// MustTrytesToTrits converts a slice of trytes into trits.
func MustTrytesToTrits(trytes Trytes) Trits {
	trits, err := TrytesToTrits(trytes)
	if err != nil {
		panic(err)
	}
	return trits
}

// Pad pads the given trytes with 9s up to the given size.
func Pad(trytes Trytes, size int) Trytes {
	if len(trytes) >= size {
		return trytes
	}
	out := make([]byte, size)
	copy(out, []byte(trytes))

	for i := len(trytes); i < size; i++ {
		out[i] = '9'
	}
	return Trytes(out)
}

// PadTrits pads the given trits with 0 up to the given size.
func PadTrits(trits Trits, size int) Trits {
	if len(trits) >= size {
		return trits
	}
	sized := make(Trits, size)
	for i := 0; i < size; i++ {
		if len(trits) > i {
			sized[i] = trits[i]
			continue
		}
		sized[i] = 0
	}
	return sized
}

func sum(a int8, b int8) int8 {
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
	sA := sum(a, b)
	cA := cons(a, b)
	cB := cons(sA, c)
	cOut := any(cA, cB)
	sOut := sum(sA, c)
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
