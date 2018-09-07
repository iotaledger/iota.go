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
	"bytes"
	"testing"
)

func TestValidTryte(t *testing.T) {
	type validTryteTC struct {
		in    rune
		valid bool
	}

	var validTryteCases = []validTryteTC{
		{in: 'A', valid: true},
		{in: 'Z', valid: true},
		{in: '9', valid: true},
		{in: '8', valid: false},
		{in: 'a', valid: false},
		{in: '-', valid: false},
		{in: 'Ɩ', valid: false},
	}

	for _, tc := range validTryteCases {
		if (IsValidTryte(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTryte(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrytes(t *testing.T) {
	type validtrytetestcase struct {
		in    Trytes
		valid bool
	}

	var validTryteCases = []validtrytetestcase{
		{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9", valid: true},
		{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ90", valid: false},
		{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9 ", valid: false},
		{in: "Ɩ", valid: false},
	}

	for _, tc := range validTryteCases {
		if (tc.in.IsValid() == nil) != tc.valid {
			t.Fatalf("ValidTrytes(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrit(t *testing.T) {
	type validtrittestcase struct {
		in    int8
		valid bool
	}

	var validTritCases = []validtrittestcase{
		{in: -1, valid: true},
		{in: 0, valid: true},
		{in: 1, valid: true},
		{in: -2, valid: false},
		{in: 2, valid: false},
	}

	for _, tc := range validTritCases {
		if (IsValidTrit(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTrit(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrits(t *testing.T) {
	type validtritstestcase struct {
		in    Trits
		valid bool
	}

	var validTritsCases = []validtritstestcase{
		{in: Trits{0}, valid: true},
		{in: Trits{-1}, valid: true},
		{in: Trits{1}, valid: true},
		{in: Trits{0, -1, 1}, valid: true},
		{in: Trits{2, -1, 1}, valid: false},
	}

	for _, tc := range validTritsCases {
		if (tc.in.IsValid() == nil) != tc.valid {
			t.Fatalf("ValidTrits(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestTritByteTrit(t *testing.T) {
	ts := []Trits{
		Trytes("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN").Trits(),
		Trytes("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM").Trits(),
		Trytes("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA").Trits(),
		Trytes("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ").Trits(),
		Trytes("999999999999999999999999999999999999999999999999999999999999999999999999999999999").Trits(),
		Trytes("SCYLJDWIM9LIXCSLETSHLQOOFDKYOVFZLAHQYCCNMYHRTNIKBZRIRACFYPOWYNSOWDNXFZUG9OEOZPOTD").Trits(),
	}

	for _, trits := range ts {
		trits[TritHashLength-1] = 0
		b, err := trits.Bytes()
		if err != nil {
			t.Errorf("Bytes() failed: %s", err)
		}

		tb, err := BytesToTrits(b)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		if !tb.Equal(trits) {
			t.Errorf("Trits->Bytes->Trits roundtrip failed\nwanted:\t%#v\ngot:\t%#v", trits, tb)
		}
	}
}

func TestAllBytes(t *testing.T) {
	for i := 0; i < 256; i++ {
		bs := bytes.Repeat([]byte{uint8(i)}, ByteLength)

		fst, err := BytesToTrits(bs)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		bs, err = fst.Bytes()
		if err != nil {
			t.Errorf("Bytes() failed: %s", err)
		}

		snd, err := BytesToTrits(bs)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		if !fst.Equal(snd) {
			t.Errorf("Bytes->Trits->Bytes->Trits roundtrip failed\nwanted:\t%#v\ngot:\t%#v", fst, snd)
		}
	}
}

func TestConvert(t *testing.T) {
	trits := Trits{0, 1, -1, 1, 1, -1, -1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	invalid := []int8{1, -1, 2, 0, 1, -1}

	if _, err := ToTrits(invalid); err == nil {
		t.Error("expected ToTrits() return an error but got none")
	}

	if _, err := ToTrytes("A_AAA"); err == nil {
		t.Error("expected ToTrytes() return an error but got none")
	}

	var target int64 = 6562317
	i := trits.Int()
	if i != target {
		t.Errorf("expected Int() to return %d but got %d", target, i)
	}

	st := Trits{0, 1, -1, 1, 1, -1, -1, 1, 1, 0, 0, 1, 0, 1, 1}
	trits2 := IntToTrits(target, 15)
	if !st.Equal(trits2) {
		t.Error("IntToTrits() is illegal.", trits2)
	}

	trits22 := IntToTrits(-1024, 7)
	if !trits22.Equal(Trits{-1, 1, 0, 1, -1, -1, -1}) {
		t.Error("IntToTrits() is illegal.")
	}

	try := st.Trytes()
	if try != "UVKIL" {
		t.Error("Int() is illegal.", try)
	}

	trits3 := try.Trits()
	if !st.Equal(trits3) {
		t.Error("Trits() is illegal.", trits3)
	}
}

func TestTrytes_Normalize(t *testing.T) {
	var bundleHash Trytes = "DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"
	no := []int8{-13, -13, -13, -13, -11, 12, 11, 7, 2, -9, -12, -6, -10, 13, 11, 3, 12, 13, -9, -11, 7, 0, 8, 6,
		11, 3, 1, 13, 13, 13, 7, 1, 2, 0, 8, -12, 10, -10, -4, 5, -9, -7, -2, -4, 5, -9, 10, -13, -12, -2, 12, -4,
		0, -11, -5, 12, -12, 7, 4, -6, -11, 3, 0, 4, 12, 7, -8, -6, 8, 0, -6, 8, -8, 11, 10, -12, 1, -8, 10, -9, -6}

	norm := bundleHash.Normalize()
	for i := range no {
		if no[i] != norm[i] {
			t.Fatal("normalization is incorrect.")
		}
	}
}

func TestTrits_Value(t *testing.T) {
	trits := Trits{1, 1, 1, 0, 0, 0, 1, -1, 1, -1, 1}
	const expected = 44482
	if trits.Value() != expected {
		t.Fatalf("expected value %d but got %d\n", expected, trits.Value())
	}
}

func TestASCIIToTrytes(t *testing.T) {
	const ascii = "IOTA"
	const utf8 = "Γιώτα"
	const expected = "SBYBCCKB"

	trytes, err := ASCIIToTrytes(ascii)
	if err != nil {
		t.Fatalf("didn't expect an error for valid ascii input but got error: %v\n", err)
	}

	for i := range trytes {
		if trytes[i] != expected[i] {
			t.Fatalf("char at %d is %v but expected %v\n", i, trytes[i], expected[i])
		}
	}

	if _, err := ASCIIToTrytes(utf8); err == nil {
		t.Fatalf("expected an error for the invalid ascii input of %v", utf8)
	}
}

func TestTrytes_ToASCII(t *testing.T) {
	const trytes = Trytes("SBYBCCKB")
	const expected = "IOTA"

	asciiVal, err := trytes.ToASCII()
	if err != nil {
		t.Fatalf("didn't expect an error for valid tryte values for ascii conversion but got error: %v\n", err)
	}

	if asciiVal != expected {
		t.Fatalf("got converted ascii value %s but expected %s\n", asciiVal, expected)
	}

	const invalidTrytes = Trytes("AAAfasds")
	const trytesWithOddLength = Trytes("AAA")

	_, err = invalidTrytes.ToASCII()
	if err == nil {
		t.Fatalf("expected an error for non convertible tryte value %s", invalidTrytes)
	}

	if err != ErrInvalidTryteCharacter {
		t.Fatalf("expected invalid tryte char error but got: %v", err)
	}

	_, err = trytesWithOddLength.ToASCII()
	if err == nil {
		t.Fatalf("expected an error for non convertible tryte value %s", trytesWithOddLength)
	}

	if err != ErrInvalidLengthForToASCIIConversion {
		t.Fatalf("expected invalid trytes length error but got: %v", err)
	}
}
