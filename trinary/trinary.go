// Package trinary provides functions for validating and converting Trits and Trytes.
package trinary

import (
	"math"
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

	byteRadix   = [5]int8{1, 3, 9, 27, 81}
	encodedZero = []int8{1, 0, 0, -1}
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

// MustAbsInt64 returns the absolute value of an int64.
func MustAbsInt64(n int64) int64 {
	if n == -1<<63 {
		panic("value out of range")
	}
	y := n >> 63       // y ← x ⟫ 63
	return (n ^ y) - y // (x ⨁ y) - y
}

func nearestGreaterMultipleOfThree(value uint64) uint64 {
	rem := value % uint64(Radix)
	if rem == 0 {
		return value
	}
	return value + uint64(Radix) - rem
}

// MinTrits returns the length of trits needed to encode the value.
func MinTrits(value int64) uint64 {
	var num uint64 = 1
	var vp uint64 = 1

	valueAbs := uint64(MustAbsInt64(value))

	for uint64(valueAbs) > vp {
		vp = vp*uint64(Radix) + 1
		num++
	}
	return num
}

// EncodedLength returns the length of trits needed to encode the value + encoding information.
func EncodedLength(value int64) uint64 {
	if value == 0 {
		return uint64(len(encodedZero))
	}
	length := nearestGreaterMultipleOfThree(MinTrits(value))

	// trits length + encoding length
	return length + MinTrits((1<<(length/uint64(Radix)))-1)
}

// IntToTrytes converts int64 to a slice of trytes.
func IntToTrytes(value int64, trytesCnt int) Trytes {
	remainder := value
	if value < 0 {
		remainder = -value
	}

	var t Trytes

	for tryte := 0; tryte < trytesCnt; tryte++ {
		idx := remainder % 27
		remainder /= 27

		if idx > 13 {
			remainder += 1
		}

		if value < 0 && idx != 0 {
			idx = 27 - idx
		}

		t += string(TryteAlphabet[idx])
	}
	return t
}

// TrytesToInt converts a slice of trytes to int64.
func TrytesToInt(t Trytes) int64 {
	var val int64

	for i := len(t) - 1; i >= 0; i-- {
		idx := strings.Index(TryteAlphabet, string(t[i]))
		if idx > 13 {
			idx = idx - 27
		}
		val = val*27 + int64(idx)
	}
	return val
}

// IntToTrits converts int64 to a slice of trits.
func IntToTrits(value int64) Trits {
	if value == 0 {
		return Trits{0}
	}

	negative := value < 0
	size := MinTrits(value)
	valueAbs := MustAbsInt64(value)

	t := make(Trits, size)

	for i := 0; i < int(size); i++ {
		if valueAbs == 0 {
			break
		}
		trit := int8((valueAbs+1)%(TrinaryRadix) - 1)
		if negative {
			trit = -trit
		}
		t[i] = trit
		valueAbs++
		valueAbs /= TrinaryRadix
	}

	return t
}

// TritsToInt converts a slice of trits into an integer and assumes little-endian notation.
func TritsToInt(t Trits) int64 {
	var val int64
	for i := len(t) - 1; i >= 0; i-- {
		val = val*3 + int64(t[i])
	}
	return val
}

// EncodeInt64 encodes an int64 as a slice of trits with encoding information.
func EncodeInt64(value int64) (t Trits, size uint64, err error) {
	size = EncodedLength(value)

	if value == 0 {
		return encodedZero, size, nil
	}

	var encoding int64 = 0
	index := 0
	length := nearestGreaterMultipleOfThree(MinTrits(MustAbsInt64(value)))
	t = make(Trits, size)
	copy(t, IntToTrits(value))

	for i := 0; i < int(length)-TrinaryRadix; i += TrinaryRadix {
		if TritsToInt(t[i:i+TrinaryRadix]) >= 0 {
			encoding |= 1 << uint(index)
			for j := 0; j < TrinaryRadix; j++ {
				t[i+j] = -t[i+j]
			}
		}
		index++
	}

	if TritsToInt(t[length-TrinaryRadix:length]) <= 0 {
		encoding |= 1 << uint(index)
		for i := 1; i < TrinaryRadix+1; i++ {
			t[int(length)-i] = -t[int(length)-i]
		}
	}

	copy(t[length:], IntToTrits(encoding))
	return t, size, nil
}

// DecodeInt64 decodes a slice of trits with encoding information as an int64.
func DecodeInt64(t Trits) (value int64, size uint64, err error) {
	numTrits := uint64(len(t))

	equal, err := TritsEqual(t[0:4], encodedZero)
	if err != nil {
		return 0, 0, err
	}

	if equal {
		return 0, EncodedLength(0), nil
	}

	value = 0
	var encodingStart uint64 = 0

	for (encodingStart < numTrits) && (TritsToInt(t[encodingStart:encodingStart+TrinaryRadix]) <= 0) {
		encodingStart += TrinaryRadix
	}

	if encodingStart >= numTrits {
		return 0, 0, errors.New("encodingStart > numTrits")
	}

	encodingStart += TrinaryRadix
	encodingLength := MinTrits((1 << (encodingStart / TrinaryRadix)) - 1)
	encoding := TritsToInt(t[encodingStart : encodingStart+encodingLength])

	// Bound checking for the lookup table
	if encodingStart/TrinaryRadix > 13 {
		return 0, 0, errors.New("encodingStart/TrinaryRadix > 13")
	}

	for i := 0; i < int(encodingStart/TrinaryRadix); i++ {
		tryteValue := TritsToInt(t[i*TrinaryRadix : (i*TrinaryRadix)+TrinaryRadix])

		if ((encoding >> uint(i)) & 1) == 1 {
			tryteValue = -tryteValue
		}
		value += Pow27LUT[i] * tryteValue
	}

	return value, encodingStart + encodingLength, nil
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

// MustTrytesToBytes is only defined for hashes (81 Trytes). It returns 48 bytes.
func MustTrytesToBytes(trytes Trytes) []byte {
	bytes, err := TrytesToBytes(trytes)
	if err != nil {
		panic(err)
	}
	return bytes
}

// BytesToTrytes converts bytes to Trytes. Returns an error if the bytes slice is not 48 in length.
func BytesToTrytes(bytes []byte, numTrytes ...int) (Trytes, error) {
	numTrits := []int{}
	if len(numTrytes) > 0 {
		numTrits = append(numTrits, numTrytes[0]*3)
	}

	trits, err := BytesToTrits(bytes, numTrits...)
	if err != nil {
		return "", err
	}

	trits = PadTrits(trits, int(nearestGreaterMultipleOfThree(uint64(len(trits)))))
	return TritsToTrytes(trits)
}

// MustBytesToTrytes converts bytes to Trytes.
func MustBytesToTrytes(bytes []byte, numTrytes ...int) Trytes {
	trytes, err := BytesToTrytes(bytes, numTrytes...)
	if err != nil {
		panic(err)
	}
	return trytes
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

// ValidTrytes returns true if t is made of valid trytes.
func ValidTrytes(trytes Trytes) error {
	if trytes == "" {
		return ErrInvalidTrytes
	}
	for _, runeVal := range trytes {
		if (runeVal < 'A' || runeVal > 'Z') && runeVal != '9' {
			return ErrInvalidTrytes
		}
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
