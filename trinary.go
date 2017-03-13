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

var (
	tryteToTritsMappings = [][]int8{
		[]int8{0, 0, 0}, []int8{1, 0, 0}, []int8{-1, 1, 0}, []int8{0, 1, 0},
		[]int8{1, 1, 0}, []int8{-1, -1, 1}, []int8{0, -1, 1}, []int8{1, -1, 1},
		[]int8{-1, 0, 1}, []int8{0, 0, 1}, []int8{1, 0, 1}, []int8{-1, 1, 1},
		[]int8{0, 1, 1}, []int8{1, 1, 1}, []int8{-1, -1, -1}, []int8{0, -1, -1},
		[]int8{1, -1, -1}, []int8{-1, 0, -1}, []int8{0, 0, -1}, []int8{1, 0, -1},
		[]int8{-1, 1, -1}, []int8{0, 1, -1}, []int8{1, 1, -1}, []int8{-1, -1, 0},
		[]int8{0, -1, 0}, []int8{1, -1, 0}, []int8{-1, 0, 0},
	}
)

//Trits is slice of int8.
//You should not use cast, insted use ToTrits
//to ensure the validity.
type Trits []int8

//ToTrits cast Trits and checks its validity.
func ToTrits(t []int8) (Trits, error) {
	tr := Trits(t)
	err := tr.IsValid()
	return tr, err
}

//IsValidTrit returns true if t is valid trit.
func IsValidTrit(t int8) error {
	if t >= -1 && t <= 1 {
		return nil
	}
	return errors.New("invalid number")
}

//IsValid returns true if ts is valid trits.
func (t Trits) IsValid() error {
	for _, tt := range t {
		if err := IsValidTrit(tt); err != nil {
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

//Int2Trits converts int64 to trits.
func Int2Trits(v int64, size int) Trits {
	tr := make(Trits, size)
	neg := false
	if v < 0 {
		v = -v
		neg = true
	}
	for i := 0; v != 0 && i < size; i++ {
		tr[i] = int8((v+1)%3) - 1
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
//This panics if len(t)%3!=0
func (t Trits) Trytes() Trytes {
	if len(t)%3 != 0 {
		panic("length of trits must be x3.")
	}
	o := ""
	for i := 0; i < len(t)/3; i++ {
		j := t[i*3] + t[i*3+1]*3 + t[i*3+2]*9
		if j < 0 {
			j += int8(len(TryteAlphabet))
		}
		o += TryteAlphabet[j : j+1]
	}
	return Trytes(o)
}

//Trytes is string of trytes.
//You should not use cast, insted use ToTrytes
//to ensure the validity.
type Trytes string

//ToTrytes cast Trytes and checks its validity.
func ToTrytes(t string) (Trytes, error) {
	tr := Trytes(t)
	err := tr.IsValid()
	return tr, err
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
