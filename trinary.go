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

package giota

import (
	"errors"
	"fmt"
	"strings"
)

//Trit is trit type for iota.
type Trit int8

//Tryte is tryte type for iota.
type Tryte byte

//Trits is slice of Trit
type Trits []Trit

//Trytes is string of trytes.
type Trytes string

//IsValidTryte returns true if t is valid tryte.
func IsValidTryte(t rune) error {
	if ('A' <= t && t <= 'Z') || t == '9' {
		return nil
	}
	return errors.New("invalid character")
}

//IsValid returns true if st is valid trytes.
func (t Trytes) IsValid() error {
	for _, t := range t {
		if err := IsValidTryte(t); err != nil {
			return fmt.Errorf("%s in trytes", err)
		}
	}
	return nil
}

//IsValid returns true if t is valid trit.
func (t Trit) IsValid() error {
	if t >= MinTritValue && t <= MaxTritValue {
		return nil
	}
	return errors.New("invalid number")
}

//IsValid returns true if ts is valid trits.
func (t Trits) IsValid() error {
	for _, tt := range t {
		if err := tt.IsValid(); err != nil {
			return fmt.Errorf("%s in trits", err)
		}
	}
	return nil
}

//Equal returns true if a and b are equal.
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

var (
	tryteToTritsMappings = [][]Trit{
		[]Trit{0, 0, 0}, []Trit{1, 0, 0}, []Trit{-1, 1, 0}, []Trit{0, 1, 0},
		[]Trit{1, 1, 0}, []Trit{-1, -1, 1}, []Trit{0, -1, 1}, []Trit{1, -1, 1},
		[]Trit{-1, 0, 1}, []Trit{0, 0, 1}, []Trit{1, 0, 1}, []Trit{-1, 1, 1},
		[]Trit{0, 1, 1}, []Trit{1, 1, 1}, []Trit{-1, -1, -1}, []Trit{0, -1, -1},
		[]Trit{1, -1, -1}, []Trit{-1, 0, -1}, []Trit{0, 0, -1}, []Trit{1, 0, -1},
		[]Trit{-1, 1, -1}, []Trit{0, 1, -1}, []Trit{1, 1, -1}, []Trit{-1, -1, 0},
		[]Trit{0, -1, 0}, []Trit{1, -1, 0}, []Trit{-1, 0, 0},
	}
)

//Int2Trits converts int64 to trits.
func Int2Trits(v int64, size int) Trits {
	tr := make(Trits, size)
	neg := false
	if v < 0 {
		v = -v
		neg = true
	}
	for i := 0; v != 0 && i < size; i++ {
		tr[i] = Trit((v+1)%3) - 1
		if neg {
			tr[i] = -tr[i]
		}
		v = (v + 1) / 3
	}
	return tr
}

// Int takes a slice of trits and converts them into an integer,
// Assumes big-endian notation.
func (t Trits) Int() int64 {
	var val int64
	for i := len(t) - 1; i >= 0; i-- {
		val = val*3 + int64(t[i])
	}
	return val
}

// Trytes takes a slice of trits and converts them into trytes,
func (t Trits) Trytes() Trytes {
	if len(t)%3 != 0 {
		panic("length of trits must be x3.")
	}
	o := ""
	for i := 0; i < len(t)/3; i++ {
		j := t[i*3] + t[i*3+1]*3 + t[i*3+2]*9
		if j < 0 {
			j += Trit(len(TryteAlphabet))
		}
		o += TryteAlphabet[j : j+1]
	}
	return Trytes(o)
}

// Trits takes a slice of trytes and converts them into tryits,
func (t Trytes) Trits() Trits {
	trits := make(Trits, len(t)*3)
	for i := range t {
		idx := strings.Index(TryteAlphabet, string(t[i:i+1]))
		copy(trits[i*3:i*3+3], tryteToTritsMappings[idx])
	}
	return trits
}

//Normalize converts trits sum of whose bits is zero.
func (t Trytes) Normalize() []int8 {
	normalized := make([]int8, len(t))
	sum := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 27; j++ {
			normalized[i*27+j] = int8(t[i*27+j : i*27+j+1].Trits().Int())
			sum += int(normalized[i*27+j])
		}
		if sum >= 0 {
			for ; sum > 0; sum-- {
				for j := 0; j < 27; j++ {
					if normalized[i*27+j] > -13 {
						normalized[i*27+j]--
						break
					}
				}
			}
		} else {
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
